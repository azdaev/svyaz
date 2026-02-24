package repo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

func GenerateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (r *Repo) CreateSession(ctx context.Context, token string, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, time.Now().Add(30*24*time.Hour),
	)
	return err
}

func (r *Repo) GetSessionUser(ctx context.Context, token string) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM sessions WHERE token = ? AND expires_at > ?`,
		token, time.Now(),
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return userID, err
}

func (r *Repo) DeleteSession(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (r *Repo) CleanExpiredSessions(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, time.Now())
	return err
}
