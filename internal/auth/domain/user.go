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
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TokenPair представляет пару токенов доступа
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// RefreshToken представляет токен обновления
type RefreshToken struct {
	ID        string
	UserID    int64
	Token     string
	ExpiresAt time.Time
}

// UserRepository определяет методы для работы с хранилищем пользователей
type UserRepository interface {
	GetByID(id int64) (*User, error)
	GetByUsername(username string) (*User, error)
	Create(user *User) error
	Update(user *User) error
}

// TokenRepository определяет методы для работы с токенами
type TokenRepository interface {
	StoreRefreshToken(token *RefreshToken) error
	GetRefreshToken(token string) (*RefreshToken, error)
	DeleteRefreshToken(token string) error
}
