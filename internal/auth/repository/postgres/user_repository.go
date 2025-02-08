package postgres

import (
	"database/sql"
	"errors"

	"github.com/Nzyazin/zadnik.store/internal/auth/domain"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id int64) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(
		`SELECT id, username, password, created_at, updated_at 
		FROM users WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(
		`SELECT id, username, password, created_at, updated_at 
		FROM users WHERE username = $1`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.QueryRow(
		`INSERT INTO users (username, password, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at`,
		user.Username,
		user.Password,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) Update(user *domain.User) error {
	result, err := r.db.Exec(
		`UPDATE users 
		SET username = $1, password = $2, updated_at = NOW()
		WHERE id = $3`,
		user.Username,
		user.Password,
		user.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
