package config

import (
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

type Config struct {
	DB                *DBConfig
	AuthServiceAddress string
	JWTSecret         string
	LOG_FILE          string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Load() (*Config, error) {
	// Загружаем .env файл
	err := godotenv.Load(filepath.Join("internal", "auth", "config", ".env-auth"))
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
		AuthServiceAddress: os.Getenv("AUTH_SERVICE_ADDRESS"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		LOG_FILE:          os.Getenv("LOG_FILE"),
	}, nil
}
