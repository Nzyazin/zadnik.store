package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound      = errors.New("user not found")
)

// User представляет собой доменную модель пользователя
type User struct {
	ID        int64
	Username  string
	PasswordHash  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepository определяет методы для работы с хранилищем пользователей
type UserRepository interface {
	GetByID(id int64) (*User, error)
	GetByUsername(username string) (*User, error)
}
