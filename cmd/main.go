package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "app/docs"
	"app/autogen/models"
	"app/autogen/routes"
	"app/internal/observability"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// @title APSO API
// @version 1.0.0
// @description Go Gin API with Swagger documentation
// @host localhost:3000
// @BasePath /api

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry tracing. This is a no-op unless tracing is
	// explicitly enabled (see internal/observability). It never blocks startup
	// on an unreachable collector and never panics; a setup error is logged and
	// the service continues without tracing.
	shutdownTracing, err := observability.InitTracing(ctx)
	if err != nil {
		log.Printf("[otel] tracing setup failed, continuing without tracing: %v", err)
		shutdownTracing = func(context.Context) error { return nil }
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(shutdownCtx); err != nil {
			log.Printf("[otel] error flushing traces on shutdown: %v", err)
		}
	}()

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

	// Register the GORM OpenTelemetry plugin so DB queries become spans. This
	// is only registered when tracing is enabled, so the default query path is
	// untouched when tracing is off.
	if observability.Enabled() {
		if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
			log.Printf("[otel] failed to register GORM tracing plugin: %v", err)
		}
	}

	// Auto-migrate models when DATABASE_SYNC is enabled
	if os.Getenv("DATABASE_SYNC") == "true" {
		if err := db.AutoMigrate(models.All()...); err != nil {
			log.Fatalf("Failed to auto-migrate: %v", err)
		}
	}

	// Create Gin router
	r := gin.Default()

	// Wire OpenTelemetry HTTP instrumentation. Registered only when tracing is
	// enabled so the default request path is unchanged when tracing is off.
	if observability.Enabled() {
		r.Use(otelgin.Middleware(observability.ServiceName()))
	}

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

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Swagger docs available at http://localhost:%s/_docs\n", port)

	// Run the server in a goroutine so we can listen for shutdown signals.
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for an interrupt or termination signal, then shut down gracefully.
	// The deferred tracing shutdown above flushes any buffered spans afterward.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
