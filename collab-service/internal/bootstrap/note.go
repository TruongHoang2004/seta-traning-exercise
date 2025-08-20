package bootstrap

import (
	"collab-service/internal/application"
	"collab-service/internal/infrastructure/persistence"
	"collab-service/internal/interface/http/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitNoteModule(r *gin.Engine, db *gorm.DB) {
	noteRepo := persistence.NewNoteRepository(db)
	noteService := application.NewNoteService(noteRepo)
	noteHandler := handler.NewNoteHandler(noteService)

	r.POST("/notes", noteHandler.Create)
	r.GET("/notes/:id", noteHandler.GetByID)
	r.GET("/notes", noteHandler.GetAll)
	r.PUT("/notes/:id", noteHandler.Update)
	r.DELETE("/notes/:id", noteHandler.Delete)
	r.POST("/notes/:noteID/share", noteHandler.ShareNote)
	r.DELETE("/notes/:noteID/share/:userID", noteHandler.RevokeAccess)
}
