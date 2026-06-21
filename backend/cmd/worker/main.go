package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/vidarnilsson/vinlaro-chat/config"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("✓ Worker starting, consuming from topic: %s", cfg.KafkaTopicMessages)
	log.Printf("  Kafka brokers: %s", cfg.KafkaBrokers)

	// Wire up aiokafka consumer here:
	//   1. Connect to Kafka
	//   2. Subscribe to cfg.KafkaTopicMessages
	//   3. For each message: write to Postgres, broadcast to WebSocket hub

	// Graceful shutdown on CTRL+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("✓ Worker running. Press CTRL+C to stop.")
	select {
	case <-quit:
		log.Println("Shutting down worker...")
		cancel()
	case <-ctx.Done():
	}
}
