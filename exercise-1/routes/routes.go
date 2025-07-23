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
		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/teams", controllers.CreateTeam)
			protected.POST("/teams/:teamId/members", controllers.AddMemberToTeam)
			protected.DELETE("/teams/:teamId/members/:memberId", controllers.RemoveMemberFromTeam)
			protected.POST("/teams/:teamId/managers", controllers.AddManagerToTeam)
			protected.DELETE("/teams/:teamId/managers/:managerId", controllers.RemoveManagerFromTeam)
		}
	}

	// GraphQL handler với middleware JWT
	router.POST("/graphql", middleware.OptionalAuthMiddleware(), func(c *gin.Context) {
		srv := handler.NewDefaultServer(
			generated.NewExecutableSchema(
				generated.Config{
					Resolvers: &graph.Resolver{},
				},
			),
		)
		srv.ServeHTTP(c.Writer, c.Request)
	})

	// Playground UI
	router.GET("/playground", gin.WrapH(playground.Handler("GraphQL playground", "/graphql")))
}
