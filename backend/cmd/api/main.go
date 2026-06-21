package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/vidarnilsson/vinlaro-chat/config"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/handler"
	"github.com/vidarnilsson/vinlaro-chat/internal/middleware"
)

func main() {
	// Load .env file if it exists (ignored in production)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to Postgres
	conn, err := sql.Open("pgx", cfg.DBConnString())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()

	if err := conn.PingContext(context.Background()); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("✓ Connected to Postgres")

	queries := db.New(conn)

	// Handlers
	authHandler := handler.NewAuthHandler(queries, cfg)
	channelHandler := handler.NewChannelHandler(queries)

	// Router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			protected.GET("/channels", channelHandler.ListChannels)
			protected.POST("/channels", channelHandler.CreateChannel)
		}
	}

	addr := fmt.Sprintf(":%s", cfg.APIPort)
	log.Printf("✓ API server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
		os.Exit(1)
	}
}
