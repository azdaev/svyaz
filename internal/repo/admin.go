package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"svyaz/internal/models"
)

func (r *Repo) AdminListUsers(ctx context.Context, search string, limit, offset int) ([]models.User, error) {
	query := `SELECT id, tg_id, tg_username, name, bio, experience, skills, photo_url, tg_chat_id, onboarded, is_admin, is_banned, created_at, updated_at FROM users`
	var args []interface{}

	if search != "" {
		query += ` WHERE (name LIKE ? OR tg_username LIKE ?)`
		s := "%" + search + "%"
		args = append(args, s, s)
	}

	query += ` ORDER BY created_at DESC`

	if limit <= 0 {
		limit = 50
	}
	query += fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("admin list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var skillsJSON string
		if err := rows.Scan(&u.ID, &u.TgID, &u.TgUsername, &u.Name, &u.Bio, &u.Experience, &skillsJSON, &u.PhotoURL, &u.TgChatID, &u.Onboarded, &u.IsAdmin, &u.IsBanned, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(skillsJSON), &u.Skills)
		users = append(users, u)
	}
	return users, nil
}

func (r *Repo) AdminListProjects(ctx context.Context, search, statusFilter string, limit, offset int) ([]models.Project, error) {
	query := `SELECT p.id, p.slug, p.author_id, p.title, p.description, p.stack, p.status, p.created_at, p.updated_at FROM projects p`
	var args []interface{}
	var conditions []string

	if search != "" {
		conditions = append(conditions, `p.title LIKE ?`)
		args = append(args, "%"+search+"%")
	}

	if statusFilter != "" {
		conditions = append(conditions, `p.status = ?`)
		args = append(args, statusFilter)
	}

	if len(conditions) > 0 {
		query += ` WHERE ` + joinConditions(conditions)
	}

	query += ` ORDER BY p.created_at DESC`

	if limit <= 0 {
		limit = 50
	}
	query += fmt.Sprintf(` LIMIT %d OFFSET %d`, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("admin list projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var stackJSON string
		if err := rows.Scan(&p.ID, &p.Slug, &p.AuthorID, &p.Title, &p.Description, &stackJSON, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(stackJSON), &p.Stack)

		author, err := r.GetUser(ctx, p.AuthorID)
		if err != nil {
			return nil, err
		}
		p.Author = author

		projects = append(projects, p)
	}
	return projects, nil
}

func (r *Repo) SetUserAdmin(ctx context.Context, userID int64, isAdmin bool) error {
	val := 0
	if isAdmin {
		val = 1
	}
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_admin = ? WHERE id = ?`, val, userID)
	return err
}

func (r *Repo) SetUserBanned(ctx context.Context, userID int64, isBanned bool) error {
	val := 0
	if isBanned {
		val = 1
	}
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_banned = ? WHERE id = ?`, val, userID)
	return err
}

func (r *Repo) SetProjectStatus(ctx context.Context, projectID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE projects SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, status, projectID)
	return err
}

func (r *Repo) AdminDeleteUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	return err
}

func (r *Repo) AdminStats(ctx context.Context) (*models.AdminStats, error) {
	s := &models.AdminStats{}

	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&s.UserCount)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects`).Scan(&s.ProjectTotal)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects WHERE status = 'pending'`).Scan(&s.ProjectPending)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects WHERE status = 'active'`).Scan(&s.ProjectActive)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects WHERE status = 'hidden'`).Scan(&s.ProjectHidden)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM responses`).Scan(&s.ResponseCount)

	return s, nil
}

func joinConditions(conditions []string) string {
	result := conditions[0]
	for i := 1; i < len(conditions); i++ {
		result += " AND " + conditions[i]
	}
	return result
}
