package main

import (
	"fmt"
	"log"
	"os"

	_ "app/docs"
	"app/autogen/models"
	"app/autogen/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title APSO API
// @version 1.0.0
// @description Go Gin API with Swagger documentation
// @host localhost:3000
// @BasePath /api

func main() {
	// Database configuration
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=apso_db port=5432 sslmode=disable"
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// DATABASE_SYNC=true creates the schema from .apsorc, the same tables,
	// columns, indexes and constraints the migrations from `apso migrate`
	// create. It only adds what is missing. Deploys run the migrations instead.
	if os.Getenv("DATABASE_SYNC") == "true" {
		if err := models.Migrate(db); err != nil {
			log.Fatalf("Failed to sync the database schema: %v", err)
		}
	}

	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger documentation endpoint - redirect root to index.html
	r.GET("/_docs", func(c *gin.Context) {
		c.Redirect(301, "/_docs/index.html")
	})
	r.GET("/_docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register API routes
	api := r.Group("/api")
	routes.RegisterAllRoutes(api, db)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Swagger docs available at http://localhost:%s/_docs\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
