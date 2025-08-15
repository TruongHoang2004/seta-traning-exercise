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

	// Manager-only APIs
	managerOnly := api.Group("/")
	managerOnly.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("manager"))
	{
		managerOnly.POST("/import-users", controllers.ImportUsersHandler)
		managerOnly.GET("/teams/:teamId/assets", controllers.GetTeamAssets)
		managerOnly.GET("/users/:userId/assets", controllers.GetUserAssets)
	}

	// Team management routes
	teams := api.Group("/teams")
	teams.Use(middleware.AuthMiddleware())
	{
		teams.POST("", controllers.CreateTeam)
		teams.GET("/:teamId", controllers.GetTeamByID)
		teams.POST("/:teamId/members", controllers.AddMemberToTeam)
		teams.DELETE("/:teamId/members/:memberId", controllers.RemoveMemberFromTeam)
		teams.POST("/:teamId/managers", controllers.AddManagerToTeam)
		teams.DELETE("/:teamId/managers/:managerId", controllers.RemoveManagerFromTeam)
	}

	// Folder management routes
	folders := api.Group("/folders")
	folders.Use(middleware.AuthMiddleware())
	{
		folders.POST("", controllers.CreateFolder)
		folders.GET("/:folderId", controllers.GetFolder)
		folders.GET("", controllers.GetFolders)
		folders.PUT("/:folderId", controllers.UpdateFolder)
		folders.DELETE("/:folderId", controllers.DeleteFolder)
		folders.POST("/:folderId/share", controllers.ShareFolder)
		folders.DELETE("/:folderId/share/:userId", controllers.RevokeFolderShare)
	}

	// Note management routes
	notes := api.Group("/notes")
	notes.Use(middleware.AuthMiddleware())
	{
		notes.POST("", controllers.CreateNote)
		notes.GET("", controllers.GetNotes)
		notes.GET("/:noteId", controllers.GetNote)
		notes.PUT("/:noteId", controllers.UpdateNote)
		notes.DELETE("/:noteId", controllers.DeleteNote)
		notes.POST("/:noteId/share", controllers.ShareNote)
		notes.DELETE("/:noteId/share/:userId", controllers.RevokeNoteShare)
	}

	return router
}
