package bootstrap

import (
	"collab-service/internal/application"
	"collab-service/internal/infrastructure/persistence/repository"
	"collab-service/internal/interface/http/handler"
	"collab-service/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitNoteModule(r *gin.Engine, db *gorm.DB) {
	noteRepo := repository.NewNoteRepository(db)
	noteService := application.NewNoteService(noteRepo)
	noteHandler := handler.NewNoteHandler(noteService)

	noteRoutes := r.Group("/api/notes")
	noteRoutes.Use(middleware.AuthMiddleware())
	{
		noteRoutes.POST("", noteHandler.Create)
		noteRoutes.GET("/:id", noteHandler.GetByID)
		noteRoutes.GET("", noteHandler.GetAll)
		noteRoutes.PUT("/:id", noteHandler.Update)
		noteRoutes.DELETE("/:noteID", noteHandler.Delete)
		noteRoutes.POST("/:noteID/shares", noteHandler.ShareNote)
		noteRoutes.DELETE("/:noteID/shares/:userID", noteHandler.RevokeAccess)
	}
}
