package config

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/joho/godotenv"
)

type Config struct {
	StoragePath string
	ImageBaseURL string
	RabbitMQURL string
}

func LoadConfig() (*Config, error) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	if err := godotenv.Load(filepath.Join("internal", "image", "config", ".env-image")); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	return &Config{
		StoragePath: filepath.Join(projectDir, os.Getenv("STORAGE_PATH")),
		ImageBaseURL: os.Getenv("IMAGE_BASE_URL"),
		RabbitMQURL: os.Getenv("RABBITMQ_URL"),
	}, nil
}