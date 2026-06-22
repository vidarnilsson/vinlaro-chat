package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/config"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/messaging"
	"github.com/vidarnilsson/vinlaro-chat/internal/model"
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

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	consumer, err := messaging.NewConsumer(brokers, cfg.KafkaTopicMessages, "chat-worker")
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	log.Printf("✓ Worker consuming from topic: %s (group: chat-worker)", cfg.KafkaTopicMessages)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go consumer.Consume(ctx, func(event model.MessageEvent) {
		if err := persistMessage(ctx, queries, event); err != nil {
			log.Printf("worker: failed to persist message %s: %v", event.ID, err)
		}
	})

	log.Println("✓ Worker running. Press CTRL+C to stop.")
	<-quit
	log.Println("Shutting down worker...")
	cancel()
}

func persistMessage(ctx context.Context, queries *db.Queries, event model.MessageEvent) error {
	msgID, err := uuid.Parse(event.ID)
	if err != nil {
		return err
	}
	channelID, err := uuid.Parse(event.ChannelID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		return err
	}

	createdAt := event.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	_, err = queries.CreateMessageWithID(ctx, db.CreateMessageWithIDParams{
		ID:        msgID,
		ChannelID: channelID,
		UserID:    userID,
		Content:   event.Content,
		CreatedAt: createdAt,
	})
	return err
}
