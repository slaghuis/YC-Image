package router

import (
  "net/http"
  "github.com/gin-gonic/gin"
	"github.com/slaghuis/YC-Image/pkg/handler"
  "github.com/slaghuis/YC-Image/pkg/config"
  "github.com/slaghuis/YC-Image/pkg/middleware"
)

func NewRouter(h *handler.Handler, cfg config.JwtConfigurations) *gin.Engine {
    r := gin.Default()

    // Health
    r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

    v1 := r.Group("/api/v1")

    // Example protected route for your own testing
    auth := v1.Group("/")
    auth.Use(middleware.RequireAccessToken(cfg))

    auth.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": c.GetString("user_role")}) })

    auth.POST("/upload", h.UploadImageHandler)
    auth.GET("/", h.ListImagesHandler)            // ?category=product
    auth.GET("/:id", h.GetImageHandler)           // metadata only
    auth.PUT("/:id", h.UpdateImageHandler)
    auth.DELETE("/:id", h.DeleteImageHandler)
    auth.GET("/product/:productID", h.GetProductGalleryHandler)
    auth.PUT("/product/:productID/order", h.UpdateProductGalleryOrderHandler)
    auth.PUT("/makeprimary/:id", h.SetPrimaryImageHandler)
    auth.GET("/primary", h.GetPrimaryImageHandler)

    return r
}
