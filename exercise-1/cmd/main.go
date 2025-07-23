package main

import (
	"fmt"
	"os"
	"seta-training-exercise-1/database"
	"seta-training-exercise-1/routes"

	"github.com/gin-gonic/gin"

	docs "seta-training-exercise-1/docs" // đảm bảo đúng với module name

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           SETA Training API
// @version         1.0
// @description     User, Team, and Asset Management
// @host            localhost:8080
// @BasePath        /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter JWT token like: Bearer <your-token>

func main() {
	database.Connect()

	router := gin.Default()

	// Set Swagger base path
	docs.SwaggerInfo.BasePath = "/api"

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Swagger at http://localhost:%s/swagger/index.html\n", port)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.SetupRoutes(router)

	router.Run(":" + port)
}
