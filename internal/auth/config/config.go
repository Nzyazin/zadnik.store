package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPass             string
	DBName             string
	JWTSecret          string
	AuthServiceAddress string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("internal/auth/config/.env-auth")
	if err != nil {
		return nil, fmt.Errorf("unable to load .env-auth file: %w", err)
	}

	return &Config{
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             os.Getenv("DB_PORT"),
		DBUser:             os.Getenv("DB_USER"),
		DBPass:             os.Getenv("DB_PASSWORD"),
		DBName:             os.Getenv("DB_NAME"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AuthServiceAddress: os.Getenv("AUTH_SERVICE_ADDRESS"),
	}, nil
}
