package handler

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "github.com/slaghuis/YC-Image/pkg/models"
)


func(h *Handler) SetPrimaryImageHandler(c *gin.Context) {
  id := c.Param("id")
  err := models.SetPrimaryImage(h.db, id)
  if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{
          "error": "Failed to set primary image",
      })
      return
  }

  c.JSON(http.StatusOK, gin.H{
      "message": "Primary image set successfully!",
      "id":             id,
  })

}



/* GET /primary_image?category=product&page=2&limit=10
 * ***********************************************************
 * Example Response
 *
{
      "id": "c4a13fa3-7cdb-4ce4-a90d-a233f85e208e",
      "category": "product",
      "original": "/images/product/c4a1-orig.jpg",
      "main": "/images/product/c4a1-main.webp",
      "thumbnail": "/images/product/c4a1-thumb.webp",
      "reference_id": "123",
      "reference_type": "product",
      "created_at": "2026-03-05T14:00:00Z",
      "updated_at": "2026-03-05T14:00:00Z"
}
 * ****************************************************/

func (h *Handler) GetPrimaryImageHandler(c *gin.Context) {
    // Query parameters
    category := c.Query("category")
    refID := c.Query("reference_id")
    refType := c.Query("reference_type")

    // Build query
    query := h.db.Model(&models.Image{})

    if category != "" {
        query = query.Where("category = ?", category)
    }

    if refID != "" {
        query = query.Where("reference_id = ?", refID)
    }

    if refType != "" {
        query = query.Where("reference_type = ?", refType)
    }

    // Query DB
    var image models.Image
    if err := query.Order("is_primary desc").
        Limit(1).
        Find(&image).Error; err != nil {

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to retreive image",
        })
        return
    }

    c.JSON(http.StatusOK, image)
}
