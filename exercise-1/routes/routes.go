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
			// Team management routes
			protected.POST("/teams", controllers.CreateTeam)
			protected.POST("/teams/:teamId/members", controllers.AddMemberToTeam)
			protected.DELETE("/teams/:teamId/members/:memberId", controllers.RemoveMemberFromTeam)
			protected.POST("/teams/:teamId/managers", controllers.AddManagerToTeam)
			protected.DELETE("/teams/:teamId/managers/:managerId", controllers.RemoveManagerFromTeam)

			// Folder management routes
			protected.POST("/folders", controllers.CreateFolder)
			protected.GET("/folders/:folderId", controllers.GetFolder)
			protected.PUT("/folders/:folderId", controllers.UpdateFolder)
			protected.DELETE("/folders/:folderId", controllers.DeleteFolder)

			// Note management routes
			protected.POST("/notes", controllers.CreateNote)
			protected.GET("/notes/:noteId", controllers.GetNote)
			protected.PUT("/notes/:noteId", controllers.UpdateNote)
			protected.DELETE("/notes/:noteId", controllers.DeleteNote)

			// Sharing routes
			protected.POST("/folders/:folderId/share", controllers.ShareFolder)
			protected.DELETE("/folders/:folderId/share/:userId", controllers.RevokeFolderShare)
			protected.POST("/notes/:noteId/share", controllers.ShareNote)
			protected.DELETE("/notes/:noteId/share/:userId", controllers.RevokeNoteShare)

			// Manager-only APIs
			protected.GET("/teams/:teamId/assets", controllers.GetTeamAssets)
			protected.GET("/users/:userId/assets", controllers.GetUserAssets)
		}
	}

	// GraphQL handler without middleware
	router.POST("/graphql", func(c *gin.Context) {
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
