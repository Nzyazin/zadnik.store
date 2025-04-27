package config

import (
	"os"
	"path/filepath"

	"github.com/Nzyazin/zadnik.store/internal/broker"
	"github.com/joho/godotenv"
)

type Config struct {
	DB                *DBConfig
	ProductServiceAddress string
	JWTSecret         string
	APIKey            string
	RabbitMQ broker.RabbitMQConfig
	LOG_FILE string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func GetConfig() (*Config, error) {
	// Загружаем .env файл
	err := godotenv.Load(filepath.Join("internal", "product", "config", ".env-product"))
	if err != nil {
		return nil, err
	}

	return &Config{
		DB: &DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		},
		ProductServiceAddress: os.Getenv("PRODUCT_SERVICE_ADDRESS"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		APIKey:            os.Getenv("API_KEY"),
		RabbitMQ: broker.RabbitMQConfig{
			URL: os.Getenv("RABBITMQ_URL"),
		},
		LOG_FILE: os.Getenv("LOG_FILE"),
	}, nil
}
