package repo

import (
	"context"
	"database/sql"
	"fmt"
	"svyaz/internal/models"
)

func (r *Repo) CreateResponse(ctx context.Context, projectID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO responses (project_id, user_id) VALUES (?, ?)`,
		projectID, userID,
	)
	if err != nil {
		return fmt.Errorf("create response: %w", err)
	}
	return nil
}

func (r *Repo) GetResponse(ctx context.Context, id int64) (*models.Response, error) {
	resp := &models.Response{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, project_id, user_id, status, created_at FROM responses WHERE id = ?`, id,
	).Scan(&resp.ID, &resp.ProjectID, &resp.UserID, &resp.Status, &resp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *Repo) HasUserResponded(ctx context.Context, projectID, userID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM responses WHERE project_id = ? AND user_id = ?`, projectID, userID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *Repo) UpdateResponseStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE responses SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *Repo) ListProjectResponses(ctx context.Context, projectID int64) ([]models.Response, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, project_id, user_id, status, created_at FROM responses WHERE project_id = ? ORDER BY created_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []models.Response
	for rows.Next() {
		var resp models.Response
		if err := rows.Scan(&resp.ID, &resp.ProjectID, &resp.UserID, &resp.Status, &resp.CreatedAt); err != nil {
			return nil, err
		}
		user, err := r.GetUser(ctx, resp.UserID)
		if err != nil {
			return nil, err
		}
		resp.User = user
		responses = append(responses, resp)
	}
	return responses, nil
}

func (r *Repo) ListUserResponses(ctx context.Context, userID int64) ([]models.Response, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.project_id, r.user_id, r.status, r.created_at
		 FROM responses r WHERE r.user_id = ? ORDER BY r.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var responses []models.Response
	for rows.Next() {
		var resp models.Response
		if err := rows.Scan(&resp.ID, &resp.ProjectID, &resp.UserID, &resp.Status, &resp.CreatedAt); err != nil {
			return nil, err
		}
		project, err := r.GetProject(ctx, resp.ProjectID)
		if err != nil {
			return nil, err
		}
		resp.Project = project
		responses = append(responses, resp)
	}
	return responses, nil
}
