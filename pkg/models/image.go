package models

/*
Explanation of Important Fields
ID (UUID)
We use UUID instead of auto‑increment for microservices.
Category
product, banner, blog, etc.
Indexed for fast lookups.
Original / Main / Thumbnail
Stores relative paths like:
/images/product/uuid-orig.jpg
/images/product/uuid-main.webp
/images/product/uuid-thumb.webp

ReferenceID + ReferenceType (Optional but recommended)
This turns your image model into a polymorphic association:
Examples:

reference_type | reference_id | meaning
"product"      |"123"         | Product with ID 123
"blog"         |"87"          | Blog post with ID 87
"banner"       |NULL          | Global banner image

This is perfect for ecommerce systems.
Timestamps
Automatically managed by Gorm.
*/
import (
    "time"
    "gorm.io/gorm"
)

type Image struct {
    ID            string     `gorm:"primaryKey;type:uuid" json:"id"`
    Category      string     `gorm:"type:varchar(50);index" json:"category"`
    Original      string     `json:"original"`
    Main          string     `json:"main"`
    Thumbnail     string     `json:"thumbnail"`
    ReferenceID   *string    `gorm:"index" json:"reference_id,omitempty"`     // product_id, caregory_id ...
    ReferenceType *string    `gorm:"index" json:"reference_type,omitempty"`   // product, category, promotion, blog
    SortOrder     int        `gorm:"default:0" json:"sort_order"`
    IsPrimary     bool       `gorm:"default:false" json:"primary"`

    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
}


func SetPrimaryImage(db *gorm.DB, imageID string) error {
    return db.Transaction(func(tx *gorm.DB) error {
        var img Image
        if err := tx.First(&img, "id = ?", imageID).Error; err != nil {
            return err
        }

        // Clear existing primary for the same ReferenceID and ReferenceType
        if err := tx.Model(&Image{}).
            Where("reference_id = ? AND reference_type = ?", img.ReferenceID, img.ReferenceType).
            Update("is_primary", false).Error; err != nil {
            return err
        }

        // Set the chosen image as primary
        if err := tx.Model(&Image{}).
            Where("id = ?", imageID).
            Update("is_primary", true).Error; err != nil {
            return err
        }

        return nil
    })
}
