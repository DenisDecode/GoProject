package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"expense-tracker/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create сохраняет пользователя. В БД попадает только хэш пароля.
func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`, email, passwordHash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, models.ErrDuplicateEmail
		}
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.getOne(ctx, `SELECT id, email, password_hash, created_at FROM users WHERE email = ?`, email)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return r.getOne(ctx, `SELECT id, email, password_hash, created_at FROM users WHERE id = ?`, id)
}

func (r *UserRepository) getOne(ctx context.Context, query string, arg any) (*models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx, query, arg).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
