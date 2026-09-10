package handler

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/h2non/bimg"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "github.com/slaghuis/YC-Image/pkg/models"
)

var allowedCategories = map[string]bool{
    "product": true,
    "banner":  true,
    "blog":    true,
    "user":    true,
}

var allowedExtensions = map[string]bool{
    ".jpg":  true,
    ".jpeg": true,
    ".png":  true,
    ".webp": true,
}


// Targets for resizing
var resizeTargets = map[string]int{
    "product": 1200,
    "banner":  1920,
    "blog":    1000,
    "user":    512,
}

func (h *Handler) UploadImageHandler(c *gin.Context) {
    // Validate category
    category := c.PostForm("category")
    if category == "" || !allowedCategories[category] {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid or missing category. Allowed: product, banner, blog, user",
        })
        return
    }

    // Parse file
    file, header, err := c.Request.FormFile("image")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
        return
    }
    defer file.Close()

    // Validate extension
    ext := strings.ToLower(filepath.Ext(header.Filename))
    if !allowedExtensions[ext] {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": fmt.Sprintf("Invalid file type: %s", ext),
        })
        return
    }

    // Create folder /images/<category>
    saveDir := fmt.Sprintf("images/%s", category)
    if err := os.MkdirAll(saveDir, 0755); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
        return
    }

    // UUID file name
    id := uuid.New().String()
    origName := id + "-orig" + ext
    mainName := id + "-main.webp"
    thumbName := id + "-thumb.webp"

    origPath := filepath.Join(saveDir, origName)
    mainPath := filepath.Join(saveDir, mainName)
    thumbPath := filepath.Join(saveDir, thumbName)

    // Save original
    origFile, err := os.Create(origPath)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot save original file"})
        return
    }
    defer origFile.Close()

    fileBytes, err := io.ReadAll(file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read uploaded file"})
        return
    }

    if _, err := origFile.Write(fileBytes); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
        return
    }

    // Process MAIN image (resize)
    targetSize := resizeTargets[category]
    mainBuf, err := bimg.NewImage(fileBytes).Process(bimg.Options{
        Width:        targetSize,
        Height:       0,
        Quality:      85,
        StripMetadata: true,
        Embed:        true,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process main image"})
        return
    }
    if err := os.WriteFile(mainPath, mainBuf, 0644); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot write main image"})
        return
    }

    // Process THUMBNAIL (square crop 300x300)
    thumbBuf, err := bimg.NewImage(fileBytes).Process(bimg.Options{
        Width:        300,
        Height:       300,
        Crop:         true,
        Quality:      80,
        StripMetadata: true,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thumbnail"})
        return
    }
    if err := os.WriteFile(thumbPath, thumbBuf, 0644); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot write thumbnail image"})
        return
    }

    refID := c.PostForm("reference_id")
    refType := c.PostForm("reference_type")

    if refType != "" && refType != "product" && refType != "blog" {
      c.JSON(http.StatusBadRequest, gin.H{
          "error": "reference_type must be 'product' or 'blog'",
        })
        return
      }

      img := models.Image{
        ID:        id,
        Category:  category,
        Original:  fmt.Sprintf("/images/%s/%s", category, origName),
        Main:      fmt.Sprintf("/images/%s/%s", category, mainName),
        Thumbnail: fmt.Sprintf("/images/%s/%s", category, thumbName),
      }

      if refID != "" {
        img.ReferenceID = &refID
      }
      if refType != "" {
        img.ReferenceType = &refType
      }

      if err := h.db.Create(&img).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image metadata"})
        return
      }

    // Return metadata + relative paths
    c.JSON(http.StatusOK, gin.H{
        "id":        id,
        "category":  category,
        "original":  fmt.Sprintf("/images/%s/%s", category, origName),
        "main":      fmt.Sprintf("/images/%s/%s", category, mainName),
        "thumbnail": fmt.Sprintf("/images/%s/%s", category, thumbName),
        "message":   "upload + resize successful",
    })
}


/* GET /images?category=product&page=2&limit=10
 * GET /images?reference_type=primar&reference_id=123
 * GET /images?category=banner
 * ***********************************************************
 * Example Response
 *
{
  "page": 1,
  "limit": 20,
  "total": 5,
  "images": [
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
  ]
}
 * ****************************************************/

func (h *Handler) ListImagesHandler(c *gin.Context) {
    // Query parameters
    category := c.Query("category")
    refID := c.Query("reference_id")
    refType := c.Query("reference_type")

    // Pagination defaults
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    if page < 1 {
        page = 1
    }
    offset := (page - 1) * limit

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
    var images []models.Image
    if err := query.Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&images).Error; err != nil {

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to list images",
        })
        return
    }

    // Count total for pagination
    var total int64
    query.Count(&total)

    c.JSON(http.StatusOK, gin.H{
        "page":  page,
        "limit": limit,
        "total": total,
        "images": images,
    })
}

func (h *Handler) UpdateImageHandler(c *gin.Context) {
    id := c.Param("id")

    // Fetch existing record
    var img models.Image
    if err := h.db.First(&img, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
        return
    }

    // Parse new uploaded file
    file, header, err := c.Request.FormFile("image")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
        return
    }
    defer file.Close()

    ext := strings.ToLower(filepath.Ext(header.Filename))
    if !allowedExtensions[ext] {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": fmt.Sprintf("Invalid file type: %s", ext),
        })
        return
    }

    // Category remains the same unless provided
    category := img.Category
    if newCat := c.PostForm("category"); newCat != "" {
        if !allowedCategories[newCat] {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Invalid category",
            })
            return
        }
        category = newCat
    }

    // Read entire file for libvips
    fileBytes, err := io.ReadAll(file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read uploaded file"})
        return
    }

    // Build new file paths
    saveDir := fmt.Sprintf("images/%s", category)
    os.MkdirAll(saveDir, 0755)

    newOrig := fmt.Sprintf("%s-orig%s", id, ext)
    newMain := fmt.Sprintf("%s-main.webp", id)
    newThumb := fmt.Sprintf("%s-thumb.webp", id)

    origPath := filepath.Join(saveDir, newOrig)
    mainPath := filepath.Join(saveDir, newMain)
    thumbPath := filepath.Join(saveDir, newThumb)

    // ----- SAVE ORIGINAL -----
    if err := os.WriteFile(origPath, fileBytes, 0644); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save original"})
        return
    }

    // ----- PROCESS MAIN (resize) -----
    target := resizeTargets[category]
    mainBuf, err := bimg.NewImage(fileBytes).Process(bimg.Options{
        Width:         target,
        Quality:       85,
        StripMetadata: true,
        Embed:         true,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process main image"})
        return
    }
    if err := os.WriteFile(mainPath, mainBuf, 0644); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot write main image"})
        return
    }

    // ----- PROCESS THUMBNAIL (300x300 crop) -----
    thumbBuf, err := bimg.NewImage(fileBytes).Process(bimg.Options{
        Width:         300,
        Height:        300,
        Crop:          true,
        Quality:       80,
        StripMetadata: true,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thumbnail"})
        return
    }
    if err := os.WriteFile(thumbPath, thumbBuf, 0644); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot write thumbnail"})
        return
    }

    // ----- DELETE OLD FILES -----
    oldPaths := []string{
        "." + img.Original,
        "." + img.Main,
        "." + img.Thumbnail,
    }
    for _, p := range oldPaths {
        if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
            fmt.Println("Warning: failed to delete old file:", p)
        }
    }

    // ----- UPDATE DB -----
    img.Original = fmt.Sprintf("/images/%s/%s", category, newOrig)
    img.Main = fmt.Sprintf("/images/%s/%s", category, newMain)
    img.Thumbnail = fmt.Sprintf("/images/%s/%s", category, newThumb)
    img.Category = category

    if err := h.db.Save(&img).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update DB"})
        return
    }

    // ----- RESPONSE -----
    c.JSON(http.StatusOK, gin.H{
        "message": "Image updated successfully",
        "id":      id,
        "category": img.Category,
        "original": img.Original,
        "main":     img.Main,
        "thumbnail": img.Thumbnail,
    })
}


/* **************************************************************
 * GET /images/4f3d927c-aaad-4dd8-b7a2-1e777b560298
 * ************************************************************
 * Example Response

{
  "id": "4f3d927c-aaad-4dd8-b7a2-1e777b560298",
  "category": "product",
  "original": "/images/product/4f3d-orig.jpg",
  "main": "/images/product/4f3d-main.webp",
  "thumbnail": "/images/product/4f3d-thumb.webp",
  "reference_id": "123",
  "reference_type": "product",
  "created_at": "2026-03-06T10:31:00Z",
  "updated_at": "2026-03-06T10:31:00Z"
}

 * Example Not Found

{
  "error": "Image not found",
  "id": "bad-id"
}


 * *****************************************************************/
func (h *Handler) GetImageHandler(c *gin.Context) {
    id := c.Param("id")

    var img models.Image
    if err := h.db.First(&img, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Image not found",
            "id":    id,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":          img.ID,
        "category":    img.Category,
        "original":    img.Original,
        "main":        img.Main,
        "thumbnail":   img.Thumbnail,
        "reference_id":   img.ReferenceID,
        "reference_type": img.ReferenceType,
        "primary":        img.IsPrimary,
        "created_at":  img.CreatedAt,
        "updated_at":  img.UpdatedAt,
    })
}

func (h *Handler) DeleteImageHandler(c *gin.Context) {
    id := c.Param("id")

    // Fetch image metadata from DB
    var img models.Image
    if err := h.db.Where("id = ?", id).First(&img).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
        return
    }

    // Build local filesystem paths
    paths := []string{
        "." + img.Original,   // stored like /images/product/uuid-orig.jpg
        "." + img.Main,       // /images/product/uuid-main.webp
        "." + img.Thumbnail,  // /images/product/uuid-thumb.webp
    }

    // Delete files from disk
    for _, p := range paths {
        if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": fmt.Sprintf("Failed to delete file: %s", p),
            })
            return
        }
    }

    // Delete database record
    if err := h.db.Delete(&models.Image{}, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to delete image database record",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Image deleted successfully",
        "id":      id,
        "deleted_files": gin.H{
            "original":  img.Original,
            "main":      img.Main,
            "thumbnail": img.Thumbnail,
        },
    })
}
