package bootstrap

import (
	"collab-service/internal/application"
	"collab-service/internal/infrastructure/persistence"
	"collab-service/internal/interface/http/handler"
	"collab-service/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitFolderModule(r *gin.Engine, db *gorm.DB) {
	folderRepo := persistence.NewFolderRepository(db)
	folderService := application.NewFolderService(folderRepo)
	folderHandler := handler.NewFolderHandler(folderService)

	group := r.Group("/api/folders")
	group.Use(middleware.AuthMiddleware())
	{
		group.POST("/", folderHandler.Create)
		group.GET("/:id", folderHandler.GetByID)
		group.GET("/", folderHandler.GetAllCanAccess)
		group.PUT("/:id", folderHandler.Update)
		group.DELETE("/:id", folderHandler.Delete)
		group.POST("/:folderID/share", folderHandler.ShareFolder)
		group.DELETE("/:folderID/share/:userID", folderHandler.RevokeAccess)
	}
}
