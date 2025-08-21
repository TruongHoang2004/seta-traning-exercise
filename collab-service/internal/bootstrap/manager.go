package bootstrap

import (
	"collab-service/config"
	"collab-service/internal/application"
	"collab-service/internal/domain/entity"
	"collab-service/internal/infrastructure/external/user_service"
	"collab-service/internal/infrastructure/persistence"
	"collab-service/internal/interface/http/handler"
	"collab-service/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitManagerModule(r *gin.Engine, db *gorm.DB) {
	client := user_service.NewGraphQLClient(config.GetConfig().UserServiceEndpoint)

	managerRepo := persistence.NewManagerRepository(db)
	teamRepo := persistence.NewTeamRepository(db)
	userRepo := user_service.NewUserRepository(client)
	managerService := application.NewManagerService(managerRepo, teamRepo, userRepo)
	managerHandler := handler.NewManagerHandler(managerService)

	// Set up routes with middleware
	managerRoutes := r.Group("/")
	managerRoutes.Use(middleware.AuthMiddleware())
	managerRoutes.Use(middleware.RoleMiddleware(entity.UserTypeManager))
	// Register routes
	{
		managerRoutes.GET("/teams/:teamId/assets", managerHandler.GetTeamAssets)
		managerRoutes.GET("/users/:userId/assets", managerHandler.GetUserAssets)
	}

}
