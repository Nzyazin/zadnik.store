package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Nzyazin/zadnik.store/pkg/db"
)

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error
}

type repository struct {
	db *db.Database
}

func NewRepository(db *db.Database) Repository {
	return &repository{db: db}
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, password, created_at, updated_at 
		FROM users 
		WHERE username = $1`, 
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	return &user, nil
}

func (r *repository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token, is_revoked, expires_at) 
		VALUES ($1, $2, $3, $4)`,
		token.UserID, token.Token, token.IsRevoked, token.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("could not save refresh token: %w", err)
	}
	return nil
}

func (r *repository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	var rt RefreshToken
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, is_revoked, created_at, expires_at 
		FROM refresh_tokens 
		WHERE token = $1 AND is_revoked = false AND expires_at > $2`,
		token, time.Now(),
	).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.IsRevoked, &rt.CreatedAt, &rt.ExpiresAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refresh token not found or expired: %w", err)
		}
		return nil, fmt.Errorf("error getting refresh token: %w", err)
	}
	return &rt, nil
}

func (r *repository) RevokeRefreshToken(ctx context.Context, token string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens 
		SET is_revoked = true 
		WHERE token = $1 AND is_revoked = false`,
		token,
	)
	if err != nil {
		return fmt.Errorf("could not revoke refresh token: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("refresh token not found or already revoked")
	}
	return nil
}

func (r *repository) DeleteExpiredTokens(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM refresh_tokens 
		WHERE expires_at < $1 OR is_revoked = true`,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("could not delete expired tokens: %w", err)
	}
	return nil
}
