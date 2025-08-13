package routes

import (
	"collab-service/internal/controllers"
	"collab-service/internal/docs"
	"collab-service/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes initializes all routes and applies necessary middleware.
func SetupRoutes() *gin.Engine {
	// Init Gin router
	router := gin.New()

	// Swagger config
	docs.SwaggerInfo.BasePath = "/api"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- Global middleware ---
	router.Use(middleware.LoggerpMiddleware()) // custom logger with zap

	api := router.Group("/api")
	{
		// Protected routes
		protected := api.Group("/")
		// protected.Use(middleware.AuthMiddleware()) // JWT Auth
		{
			// use group route for each domain
			// Team management
			protected.POST("/teams", controllers.CreateTeam)
			protected.POST("/teams/:teamId/members", controllers.AddMemberToTeam)
			protected.DELETE("/teams/:teamId/members/:memberId", controllers.RemoveMemberFromTeam)
			protected.POST("/teams/:teamId/managers", controllers.AddManagerToTeam)
			protected.DELETE("/teams/:teamId/managers/:managerId", controllers.RemoveManagerFromTeam)

			// Folder management
			protected.POST("/folders", controllers.CreateFolder)
			protected.GET("/folders/:folderId", controllers.GetFolder)
			protected.PUT("/folders/:folderId", controllers.UpdateFolder)
			protected.DELETE("/folders/:folderId", controllers.DeleteFolder)

			// Note management
			protected.POST("/notes", controllers.CreateNote)
			protected.GET("/notes/:noteId", controllers.GetNote)
			protected.PUT("/notes/:noteId", controllers.UpdateNote)
			protected.DELETE("/notes/:noteId", controllers.DeleteNote)

			// Sharing
			protected.POST("/folders/:folderId/share", controllers.ShareFolder)
			protected.DELETE("/folders/:folderId/share/:userId", controllers.RevokeFolderShare)
			protected.POST("/notes/:noteId/share", controllers.ShareNote)
			protected.DELETE("/notes/:noteId/share/:userId", controllers.RevokeNoteShare)

			// Manager-only APIs
			protected.GET("/teams/:teamId/assets", controllers.GetTeamAssets)
			protected.GET("/users/:userId/assets", controllers.GetUserAssets)
		}
	}

	return router
}
