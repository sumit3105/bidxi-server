package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"bidxi/internal/config"
	"bidxi/internal/handler"
	"bidxi/internal/middleware"
	"bidxi/internal/repository"
	"bidxi/internal/service"
	"bidxi/pkg/database"
)

func main() {
	// Load configuration
	config.Load()

	// Set Gin mode
	gin.SetMode(config.AppConfig.GinMode)

	// Connect to MongoDB
	if err := database.Connect(config.AppConfig.MongoURI, config.AppConfig.MongoDBName); err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	// Ensure database indexes
	userRepo := repository.NewUserRepository()
	if err := userRepo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("⚠️  Warning: Failed to ensure indexes: %v", err)
	}

	lobbyRepo := repository.NewLobbyRepository()
	if err := lobbyRepo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("⚠️  Warning: Failed to ensure lobby indexes: %v", err)
	}

	// Initialize services (dependency injection)
	authService := service.NewAuthService(userRepo)
	lobbyService := service.NewLobbyService(lobbyRepo, userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	lobbyHandler := handler.NewLobbyHandler(lobbyService)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(config.AppConfig.RateLimitRPM)

	// Set up Gin router
	router := gin.Default()

	// CORS configuration — allow Angular dev server
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Apply global rate limiting
	router.Use(rateLimiter.Middleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "bidxi-api",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	api := router.Group("/api")

	// Public routes (no JWT required)
	authHandler.RegisterRoutes(api)

	// Protected routes (JWT required)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	authHandler.RegisterProtectedRoutes(protected)
	lobbyHandler.RegisterRoutes(protected)

	// Start the server with graceful shutdown
	addr := fmt.Sprintf(":%s", config.AppConfig.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine
	go func() {
		log.Printf("🚀 BidXI API server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown on interrupt/terminate signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
