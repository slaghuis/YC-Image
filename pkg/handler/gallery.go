package handler

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "github.com/slaghuis/YC-Image/pkg/models"
)


/* ***********************************************
{
  "product_id": "123",
  "count": 3,
  "images": [
    {
      "id": "abc1",
      "main": "/images/product/abc1-main.webp",
      "thumbnail": "/images/product/abc1-thumb.webp",
      "sort_order": 0
    },
    {
      "id": "abc2",
      "main": "/images/product/abc2-main.webp",
      "sort_order": 1
    },
    {
      "id": "abc3",
      "main": "/images/product/abc3-main.webp",
      "sort_order": 2
    }
  ]
}
 * ***************************************************/
func (h *Handler) GetProductGalleryHandler(c *gin.Context) {
    productID := c.Param("productID")

    var images []models.Image

    err := h.db.Where("reference_id = ? AND reference_type = ?", productID, "product").
        Order("sort_order ASC, created_at ASC").
        Find(&images).Error

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch product gallery",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "product_id": productID,
        "count":      len(images),
        "images":     images,
    })
}


type SortUpdate struct {
    ImageID   string `json:"image_id"`
    SortOrder int    `json:"sort_order"`
}

func (h *Handler) UpdateProductGalleryOrderHandler(c *gin.Context) {
    productID := c.Param("productID")

    var updates []SortUpdate
    if err := c.BindJSON(&updates); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
        return
    }

    for _, u := range updates {
        h.db.Model(&models.Image{}).
            Where("id = ? AND reference_id = ? AND reference_type = ?", u.ImageID, productID, "product").
            Update("sort_order", u.SortOrder)
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Sort order updated",
        "product_id": productID,
    })
}
