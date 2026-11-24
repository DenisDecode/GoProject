package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"expense-tracker/internal/models"
)

// SessionRepository хранит сессии в таблице sessions (переживают перезапуск сервера).
type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create создаёт сессию со случайным идентификатором (32 байта из crypto/rand).
func (r *SessionRepository) Create(ctx context.Context, userID int64, ttl time.Duration) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(b)
	expires := time.Now().Add(ttl)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expires.Unix())
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

// GetUserID возвращает пользователя только для НЕпросроченной сессии.
func (r *SessionRepository) GetUserID(ctx context.Context, token string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM sessions WHERE id = ? AND expires_at > ?`,
		token, time.Now().Unix()).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, models.ErrNotFound
	}
	return userID, err
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, token)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= ?`, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
