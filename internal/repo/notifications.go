package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"svyaz/internal/models"
)

func (r *Repo) CreateNotification(ctx context.Context, userID int64, ntype string, payload map[string]interface{}) error {
	data, _ := json.Marshal(payload)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notifications (user_id, type, payload) VALUES (?, ?, ?)`,
		userID, ntype, string(data),
	)
	return err
}

func (r *Repo) ListNotifications(ctx context.Context, userID int64, limit int) ([]models.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, type, payload, read, created_at FROM notifications
		 WHERE user_id = ? ORDER BY created_at DESC LIMIT ?`, userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifs []models.Notification
	for rows.Next() {
		var n models.Notification
		var payloadJSON string
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &payloadJSON, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(payloadJSON), &n.Payload)
		notifs = append(notifs, n)
	}
	return notifs, nil
}

func (r *Repo) UnreadNotificationCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read = 0`, userID,
	).Scan(&count)
	return count, err
}

func (r *Repo) MarkNotificationsRead(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read = 1 WHERE user_id = ? AND read = 0`, userID,
	)
	return err
}
