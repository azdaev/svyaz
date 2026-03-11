package repo

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"svyaz/internal/models"
)

const slugChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateSlug() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = slugChars[int(b[i])%len(slugChars)]
	}
	return string(b)
}

type ProjectFilter struct {
	RoleSlug string
	Stack    string
	Type     string
	Limit    int
	Offset   int
}

func (r *Repo) CreateProject(ctx context.Context, authorID int64, title, description, projectType string, stack []string, roleCounts map[int64]int) (string, error) {
	stackJSON, _ := json.Marshal(stack)
	slug := generateSlug()

	res, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (author_id, title, description, stack, status, type, slug) VALUES (?, ?, ?, ?, 'pending', ?, ?)`,
		authorID, title, description, string(stackJSON), projectType, slug,
	)
	if err != nil {
		return "", fmt.Errorf("insert project: %w", err)
	}
	projectID, _ := res.LastInsertId()

	for rid, count := range roleCounts {
		if count < 1 {
			count = 1
		}
		_, err = r.db.ExecContext(ctx, `INSERT INTO project_roles (project_id, role_id, count) VALUES (?, ?, ?)`, projectID, rid, count)
		if err != nil {
			return "", fmt.Errorf("insert project role: %w", err)
		}
	}

	return slug, nil
}

func (r *Repo) GetProject(ctx context.Context, id int64) (*models.Project, error) {
	p := &models.Project{}
	var stackJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, slug, author_id, title, description, stack, status, type, created_at, updated_at FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Slug, &p.AuthorID, &p.Title, &p.Description, &stackJSON, &p.Status, &p.Type, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	_ = json.Unmarshal([]byte(stackJSON), &p.Stack)

	roles, err := r.getProjectRoles(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Roles = roles

	author, err := r.GetUser(ctx, p.AuthorID)
	if err != nil {
		return nil, err
	}
	p.Author = author

	return p, nil
}

func (r *Repo) GetProjectBySlug(ctx context.Context, slug string) (*models.Project, error) {
	p := &models.Project{}
	var stackJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, slug, author_id, title, description, stack, status, type, created_at, updated_at FROM projects WHERE slug = ?`, slug,
	).Scan(&p.ID, &p.Slug, &p.AuthorID, &p.Title, &p.Description, &stackJSON, &p.Status, &p.Type, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get project by slug: %w", err)
	}
	_ = json.Unmarshal([]byte(stackJSON), &p.Stack)

	roles, err := r.getProjectRoles(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Roles = roles

	author, err := r.GetUser(ctx, p.AuthorID)
	if err != nil {
		return nil, err
	}
	p.Author = author

	return p, nil
}

func (r *Repo) UpdateProject(ctx context.Context, id int64, title, description, projectType string, stack []string, roleCounts map[int64]int) error {
	stackJSON, _ := json.Marshal(stack)

	_, err := r.db.ExecContext(ctx,
		`UPDATE projects SET title = ?, description = ?, stack = ?, type = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		title, description, string(stackJSON), projectType, id,
	)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `DELETE FROM project_roles WHERE project_id = ?`, id)
	if err != nil {
		return fmt.Errorf("clear project roles: %w", err)
	}

	for rid, count := range roleCounts {
		if count < 1 {
			count = 1
		}
		_, err = r.db.ExecContext(ctx, `INSERT INTO project_roles (project_id, role_id, count) VALUES (?, ?, ?)`, id, rid, count)
		if err != nil {
			return fmt.Errorf("insert project role: %w", err)
		}
	}

	return nil
}

func (r *Repo) DeleteProject(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	return err
}

func (r *Repo) ListProjects(ctx context.Context, f ProjectFilter) ([]models.Project, error) {
	query := `SELECT DISTINCT p.id, p.slug, p.author_id, p.title, p.description, p.stack, p.status, p.type, p.created_at, p.updated_at FROM projects p`
	var args []interface{}
	var conditions []string

	if f.RoleSlug != "" {
		query += ` JOIN project_roles pr ON pr.project_id = p.id JOIN roles rl ON rl.id = pr.role_id`
		conditions = append(conditions, `rl.slug = ?`)
		args = append(args, f.RoleSlug)
	}

	if f.Stack != "" {
		conditions = append(conditions, `p.stack LIKE ?`)
		args = append(args, `%"`+f.Stack+`"%`)
	}

	if f.Type != "" {
		conditions = append(conditions, `p.type = ?`)
		args = append(args, f.Type)
	}

	conditions = append(conditions, `p.status = 'active'`)

	if len(conditions) > 0 {
		query += ` WHERE ` + strings.Join(conditions, ` AND `)
	}

	query += ` ORDER BY p.created_at DESC`

	if f.Limit <= 0 {
		f.Limit = 20
	}
	query += fmt.Sprintf(` LIMIT %d OFFSET %d`, f.Limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var stackJSON string
		if err := rows.Scan(&p.ID, &p.Slug, &p.AuthorID, &p.Title, &p.Description, &stackJSON, &p.Status, &p.Type, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(stackJSON), &p.Stack)

		roles, err := r.getProjectRoles(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.Roles = roles

		author, err := r.GetUser(ctx, p.AuthorID)
		if err != nil {
			return nil, err
		}
		p.Author = author

		projects = append(projects, p)
	}

	return projects, nil
}

func (r *Repo) ListUserProjects(ctx context.Context, userID int64) ([]models.Project, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, slug, author_id, title, description, stack, status, type, created_at, updated_at
		 FROM projects WHERE author_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var stackJSON string
		if err := rows.Scan(&p.ID, &p.Slug, &p.AuthorID, &p.Title, &p.Description, &stackJSON, &p.Status, &p.Type, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(stackJSON), &p.Stack)

		roles, _ := r.getProjectRoles(ctx, p.ID)
		p.Roles = roles

		projects = append(projects, p)
	}
	return projects, nil
}

func (r *Repo) getProjectRoles(ctx context.Context, projectID int64) ([]models.Role, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.slug, r.name, pr.count FROM roles r
		 JOIN project_roles pr ON pr.role_id = r.id
		 WHERE pr.project_id = ?`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Slug, &role.Name, &role.Count); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *Repo) GetProjectRolesWithFilled(ctx context.Context, projectID int64) ([]models.Role, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.slug, r.name, pr.count,
			COALESCE((SELECT COUNT(*) FROM responses resp
				  WHERE resp.project_id = pr.project_id
				    AND resp.role_id = r.id
				    AND resp.status = 'accepted'), 0) AS filled
		 FROM roles r
		 JOIN project_roles pr ON pr.role_id = r.id
		 WHERE pr.project_id = ?`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Slug, &role.Name, &role.Count, &role.Filled); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *Repo) CountProjectResponses(ctx context.Context, projectID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM responses WHERE project_id = ?`, projectID).Scan(&count)
	return count, err
}

func (r *Repo) BackfillSlugs(ctx context.Context) error {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM projects WHERE slug = ''`)
	if err != nil {
		return fmt.Errorf("backfill slugs query: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}

	for _, id := range ids {
		slug := generateSlug()
		if _, err := r.db.ExecContext(ctx, `UPDATE projects SET slug = ? WHERE id = ?`, slug, id); err != nil {
			return fmt.Errorf("backfill slug for project %d: %w", id, err)
		}
	}

	if len(ids) > 0 {
		fmt.Printf("Backfilled slugs for %d projects\n", len(ids))
	}

	return nil
}
