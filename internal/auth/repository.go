package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Nzyazin/zadnik.store/pkg/db"
)

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	SaveToken(ctx context.Context, token *Token) error
	GetUserByToken(ctx context.Context, token string) (*User, error)
}

type repository struct {
	db *db.Database
}

func NewRepository(db *db.Database) Repository {
	return &repository{db: db}
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, password, role FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) SaveToken(ctx context.Context, token *Token) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO tokens (user_id, token) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET token = $2", token.UserID, token.Token)
	if err != nil {
		return fmt.Errorf("could not save token: %w", err)
	}
	return err
}

func (r *repository) GetUserByToken(ctx context.Context, token string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.password, u.role
		FROM users u
		JOIN tokens t ON u.id = t.user_id
		WHERE t.token = $1`).
		Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}
