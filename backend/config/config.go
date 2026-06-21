package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// Kafka
	KafkaBrokers       string
	KafkaTopicMessages string

	// MinIO
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
	MinioBucket    string

	// Server
	APIPort string
}

func Load() (*Config, error) {
	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}

	minioSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))

	return &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "chat"),
		DBPassword:         getEnv("DB_PASSWORD", "chat"),
		DBName:             getEnv("DB_NAME", "chat"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me"),
		JWTExpiryHours:     jwtExpiry,
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopicMessages: getEnv("KAFKA_TOPIC_MESSAGES", "chat.messages"),
		MinioEndpoint:      getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:     getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioUseSSL:        minioSSL,
		MinioBucket:        getEnv("MINIO_BUCKET", "chat-attachments"),
		APIPort:            getEnv("API_PORT", "8000"),
	}, nil
}

func (c *Config) DBConnString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
