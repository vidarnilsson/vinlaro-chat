package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/vidarnilsson/vinlaro-chat/config"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/handler"
	"github.com/vidarnilsson/vinlaro-chat/internal/messaging"
	"github.com/vidarnilsson/vinlaro-chat/internal/middleware"
	"github.com/vidarnilsson/vinlaro-chat/internal/model"
	"github.com/vidarnilsson/vinlaro-chat/internal/ws"
)

func main() {
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

	// Kafka producer
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	producer, err := messaging.NewProducer(brokers, cfg.KafkaTopicMessages)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()
	log.Println("✓ Kafka producer ready")

	// WebSocket hub
	hub := ws.NewHub()
	go hub.Run()
	log.Println("✓ WebSocket hub running")

	// API-side Kafka consumer: broadcasts saved messages to WebSocket clients.
	apiConsumer, err := messaging.NewConsumer(brokers, cfg.KafkaTopicMessages, "chat-api")
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}
	defer apiConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go apiConsumer.Consume(ctx, func(event model.MessageEvent) {
		payload, err := json.Marshal(event)
		if err != nil {
			log.Printf("ws broadcast: marshal error: %v", err)
			return
		}
		channelID, err := uuid.Parse(event.ChannelID)
		if err != nil {
			log.Printf("ws broadcast: invalid channel id: %v", err)
			return
		}
		hub.BroadcastToChannel(channelID, payload)
	})
	log.Println("✓ API Kafka consumer running (group: chat-api)")

	// Handlers
	authHandler := handler.NewAuthHandler(queries)
	channelHandler := handler.NewChannelHandler(queries)
	messageHandler := handler.NewMessageHandler(queries, producer)
	wsHandler := handler.NewWSHandler(hub, queries)
	dmHandler := handler.NewDMHandler(queries, conn)
	userHandler := handler.NewUserHandler(queries)
	friendsHandler := handler.NewFriendsHandler(queries)
	inviteHandler := handler.NewInviteHandler(queries)

	// Router
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// WebSocket endpoint — auth via session cookie sent automatically by browser.
	r.GET("/ws/channels/:id", wsHandler.ServeWS)

	api := r.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
		}

		protected := api.Group("/")
		protected.Use(middleware.Auth(queries))
		{
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.Me)
			protected.GET("/channels", channelHandler.ListChannels)
			protected.POST("/channels", channelHandler.CreateChannel)
			protected.POST("/channels/:id/messages", messageHandler.SendMessage)
			protected.GET("/channels/:id/messages", messageHandler.GetMessages)
			protected.POST("/channels/:id/invite/:userID", inviteHandler.SendInvite)

			protected.GET("/dm", dmHandler.ListDMs)
			protected.POST("/dm/:userID", dmHandler.GetOrCreateDM)

			protected.GET("/users", userHandler.SearchUsers)

			protected.GET("/friends", friendsHandler.ListFriends)
			protected.GET("/friends/requests", friendsHandler.ListPendingRequests)
			protected.POST("/friends/request/:userID", friendsHandler.SendFriendRequest)
			protected.POST("/friends/accept/:friendshipID", friendsHandler.AcceptFriendRequest)
			protected.POST("/friends/decline/:friendshipID", friendsHandler.DeclineFriendRequest)
			protected.POST("/friends/block/:userID", friendsHandler.BlockUser)

			protected.GET("/invites", inviteHandler.ListPendingInvites)
			protected.POST("/invites/:inviteID/accept", inviteHandler.AcceptInvite)
			protected.POST("/invites/:inviteID/decline", inviteHandler.DeclineInvite)
		}
	}

	addr := fmt.Sprintf(":%s", cfg.APIPort)
	log.Printf("✓ API server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
		os.Exit(1)
	}
}
