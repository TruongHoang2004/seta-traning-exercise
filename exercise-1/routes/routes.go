package routes

import (
	"seta-training-exercise-1/controllers"
	"seta-training-exercise-1/graph"
	"seta-training-exercise-1/graph/generated"
	"seta-training-exercise-1/middleware"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.POST("/login", controllers.Login)
		api.POST("/users", controllers.CreateUser)

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/users", controllers.FetchUsers)
			protected.POST("/teams", controllers.CreateTeam)
			// Thêm các route khác sau
		}
	}

	// GraphQL handler
	router.POST("/graphql", gin.WrapH(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))))

	// Optional: Playground UI
	router.GET("/playground", gin.WrapH(playground.Handler("GraphQL playground", "/graphql")))
}
