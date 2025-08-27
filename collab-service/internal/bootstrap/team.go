package bootstrap

import (
	"collab-service/config"
	"collab-service/internal/application"
	"collab-service/internal/infrastructure/external/cache"
	"collab-service/internal/infrastructure/external/user_service"
	"collab-service/internal/infrastructure/persistence"
	"collab-service/internal/interface/http/handler"
	"collab-service/internal/interface/http/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitTeamModule(r *gin.Engine, db *gorm.DB) {
	client := user_service.NewGraphQLClient(config.GetConfig().UserServiceEndpoint)

	// cacheService := cache.NewCacheService(rdb)
	teamRepo := persistence.NewTeamRepository(db)
	teamRepoWithCache := cache.NewTeamRepositoryWithCache(teamRepo, cache.GetRedisClient(), time.Hour)
	userRepo := user_service.NewUserRepository(client)

	service := application.NewTeamService(teamRepoWithCache, userRepo)
	h := handler.NewTeamHandler(service)

	group := r.Group("/api/teams")
	group.Use(middleware.AuthMiddleware())
	{
		group.POST("", h.CreateTeam)
		group.GET(":id", h.GetByID)
		group.GET("", h.GetAllByUserID)
		group.PUT("/:teamId/members", h.AddMembers)
		group.PUT("/:teamId/managers/:managerId", h.AddManager)
		group.DELETE("/:teamId/members/:memberId", h.RemoveMember)
		group.DELETE("/:teamId/managers/:managerId", h.RemoveManager)
		group.PUT("/:teamId", h.UpdateTeam)
		group.DELETE("/:teamId", h.DeleteTeam)
	}

}
