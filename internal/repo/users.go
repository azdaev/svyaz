package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"svyaz/internal/models"
)

func (r *Repo) UpsertUser(ctx context.Context, tgID int64, tgUsername, name string) (user *models.User, isNew bool, err error) {
	var id int64
	err = r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE tg_id = ?`, tgID).Scan(&id)
	if err == sql.ErrNoRows {
		res, err := r.db.ExecContext(ctx,
			`INSERT INTO users (tg_id, tg_username, name) VALUES (?, ?, ?)`,
			tgID, tgUsername, name,
		)
		if err != nil {
			return nil, false, fmt.Errorf("insert user: %w", err)
		}
		id, _ = res.LastInsertId()
		isNew = true
	} else if err != nil {
		return nil, false, fmt.Errorf("query user: %w", err)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE users SET tg_username = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			tgUsername, id,
		)
		if err != nil {
			return nil, false, fmt.Errorf("update user: %w", err)
		}
	}

	user, err = r.GetUser(ctx, id)
	return user, isNew, err
}

func (r *Repo) GetUser(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	var skillsJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tg_id, tg_username, name, bio, experience, skills, onboarded, is_admin, is_banned, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.TgID, &u.TgUsername, &u.Name, &u.Bio, &u.Experience, &skillsJSON, &u.Onboarded, &u.IsAdmin, &u.IsBanned, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	_ = json.Unmarshal([]byte(skillsJSON), &u.Skills)

	roles, err := r.getUserRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	return u, nil
}

func (r *Repo) GetUserByTgID(ctx context.Context, tgID int64) (*models.User, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE tg_id = ?`, tgID).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetUser(ctx, id)
}

func (r *Repo) UpdateUserProfile(ctx context.Context, userID int64, name, bio, experience string, skills []string, roleIDs []int64) error {
	skillsJSON, _ := json.Marshal(skills)

	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET name = ?, bio = ?, experience = ?, skills = ?, onboarded = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, bio, experience, string(skillsJSON), userID,
	)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("clear roles: %w", err)
	}

	for _, rid := range roleIDs {
		_, err = r.db.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, userID, rid)
		if err != nil {
			return fmt.Errorf("insert role: %w", err)
		}
	}

	return nil
}

func (r *Repo) getUserRoles(ctx context.Context, userID int64) ([]models.Role, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.slug, r.name FROM roles r
		 JOIN user_roles ur ON ur.role_id = r.id
		 WHERE ur.user_id = ?`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Slug, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *Repo) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, slug, name FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Slug, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}
