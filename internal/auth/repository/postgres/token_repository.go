package postgres

import (
	"database/sql"
	"errors"

	"github.com/Nzyazin/zadnik.store/internal/auth/domain"
)

type tokenRepository struct {
	db *sql.DB
}

// NewTokenRepository создает новый экземпляр TokenRepository
func NewTokenRepository(db *sql.DB) domain.TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) StoreRefreshToken(token *domain.RefreshToken) error {
	_, err := r.db.Exec(
		`INSERT INTO refresh_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	)
	return err
}

func (r *tokenRepository) GetRefreshToken(token string) (*domain.RefreshToken, error) {
	refreshToken := &domain.RefreshToken{}
	err := r.db.QueryRow(
		`SELECT id, user_id, token, expires_at 
		FROM refresh_tokens WHERE token = $1`,
		token,
	).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	return refreshToken, nil
}

func (r *tokenRepository) DeleteRefreshToken(token string) error {
	result, err := r.db.Exec(
		"DELETE FROM refresh_tokens WHERE token = $1",
		token,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("refresh token not found")
	}

	return nil
}
