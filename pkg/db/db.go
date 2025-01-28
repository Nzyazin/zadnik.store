package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Nzyazin/zadnik.store/internal/auth/config"
	_ "github.com/lib/pq"
	"time"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Database struct {
	*sql.DB
}

func NewFromAuthConfig(authConfig *config.Config) (*Database, error) {
	cfg := Config{
		Host:     authConfig.DBHost,
		Port:     authConfig.DBPort,
		User:     authConfig.DBUser,
		Password: authConfig.DBPass,
		DBName:   authConfig.DBName,
		SSLMode:  "disable",
	}

	return New(cfg)
}

func New(cfg Config) (*Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Database{db}, nil
}

func (d *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.DB.QueryRowContext(ctx, query, args...)
}

func (d *Database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.DB.ExecContext(ctx, query, args...)
}

func (d *Database) Close() error {
	return d.DB.Close()
}
