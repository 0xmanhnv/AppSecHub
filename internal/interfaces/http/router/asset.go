package router

import (
	"appsechub/internal/interfaces/http/handler"

	"github.com/gin-gonic/gin"
)

// RegisterAssetRoutes mounts asset-related routes under /v1.
// Mirrors naming like registerAdminRoutes/registerAuthRoutes for consistency.
func RegisterAssetRoutes(v1 *gin.RouterGroup, h *handler.AssetHandler, authMiddleware ...gin.HandlerFunc) {
	assets := v1.Group("/assets")
	if len(authMiddleware) > 0 {
		assets.Use(authMiddleware...)
	}
	assets.POST("", h.UpsertAsset)
	assets.GET("", h.ListAssets)

	apps := v1.Group("/applications")
	if len(authMiddleware) > 0 {
		apps.Use(authMiddleware...)
	}
	apps.POST("", h.UpsertApplication)
	apps.POST("/:id/repos", h.LinkApplicationRepo)
}
