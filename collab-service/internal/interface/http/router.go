package http

import (
	"collab-service/docs"
	"collab-service/internal/bootstrap"
	"collab-service/internal/infrastructure/database"
	"collab-service/internal/interface/http/middleware"

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

	// Initialize modules and attach them to the /api route group
	bootstrap.InitTeamModule(router, database.GetDB())
	bootstrap.InitFolderModule(router, database.GetDB())
	bootstrap.InitNoteModule(router, database.GetDB())

	// Redirect root to /api path
	router.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/api")
	})

	return router
}
