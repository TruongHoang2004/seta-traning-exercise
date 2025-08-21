package bootstrap

import (
	"collab-service/config"
	"collab-service/internal/application"
	"collab-service/internal/domain/entity"
	"collab-service/internal/infrastructure/external/user_service"
	"collab-service/internal/interface/http/handler"
	"collab-service/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
)

func InitUserModule(r *gin.Engine) {

	client := user_service.NewGraphQLClient(config.GetConfig().UserServiceEndpoint)
	userRepo := user_service.NewUserRepository(client)

	service := application.NewUserService(userRepo)
	h := handler.NewUserHandler(service)

	group := r.Group("/api/users")
	group.Use(middleware.AuthMiddleware())
	group.Use(middleware.RoleMiddleware(entity.UserTypeManager))
	{
		group.POST("import", h.ImportUsersFromCSV)
	}

}
