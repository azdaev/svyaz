package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"svyaz/internal/middleware"
	"svyaz/internal/models"
	"svyaz/internal/repo"
)

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	roleSlug := r.URL.Query().Get("role")
	stack := r.URL.Query().Get("stack")

	projects, err := h.repo.ListProjects(r.Context(), repo.ProjectFilter{
		RoleSlug: roleSlug,
		Stack:    stack,
		Limit:    50,
	})
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	roles, _ := h.repo.GetAllRoles(r.Context())

	h.render(w, r, "index.html", map[string]any{
		"Projects":    projects,
		"Roles":       roles,
		"FilterRole":  roleSlug,
		"FilterStack": stack,
	})
}

func (h *Handler) handleProjectView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	project, err := h.repo.GetProject(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.UserFromContext(r.Context())

	// Non-active projects: only visible to author and admins
	if project.Status != "active" {
		isAuthor := user != nil && user.ID == project.AuthorID
		isAdmin := user != nil && user.IsAdmin
		if !isAuthor && !isAdmin {
			http.NotFound(w, r)
			return
		}
	}

	data := map[string]any{
		"Project":     project,
		"JustCreated": r.URL.Query().Get("created") == "1",
	}

	if user != nil {
		data["IsAuthor"] = user.ID == project.AuthorID

		if user.ID != project.AuthorID {
			if resp, err := h.repo.GetUserResponseForProject(r.Context(), id, user.ID); err == nil {
				data["HasResponded"] = true
				data["UserResponseStatus"] = resp.Status
			}
		}

		if user.ID == project.AuthorID {
			responses, _ := h.repo.ListProjectResponses(r.Context(), id)
			data["Responses"] = responses
		}
	}

	h.render(w, r, "project_view.html", data)
}

func (h *Handler) handleProjectNew(w http.ResponseWriter, r *http.Request) {
	roles, _ := h.repo.GetAllRoles(r.Context())
	h.render(w, r, "project_form.html", map[string]any{
		"Roles":  roles,
		"IsEdit": false,
	})
}

func (h *Handler) handleProjectEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	project, err := h.repo.GetProject(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user.ID != project.AuthorID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	roles, _ := h.repo.GetAllRoles(r.Context())
	roleIDs := extractRoleIDs(project.Roles)
	roleCountMap := make(map[int64]int)
	for _, role := range project.Roles {
		roleCountMap[role.ID] = role.Count
	}

	h.render(w, r, "project_form.html", map[string]any{
		"Roles":         roles,
		"IsEdit":        true,
		"Project":       project,
		"ProjectRoles":  roleIDs,
		"RoleCountMap":  roleCountMap,
	})
}

func (h *Handler) handleUserProfile(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	profile, err := h.repo.GetUser(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.render(w, r, "user.html", map[string]any{
		"Profile": profile,
	})
}

func (h *Handler) handleOnboarding(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user.Onboarded {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	roles, _ := h.repo.GetAllRoles(r.Context())
	h.render(w, r, "onboarding.html", map[string]any{
		"Roles": roles,
	})
}

func (h *Handler) handleSettings(w http.ResponseWriter, r *http.Request) {
	roles, _ := h.repo.GetAllRoles(r.Context())
	user := middleware.UserFromContext(r.Context())
	roleIDs := extractRoleIDs(user.Roles)

	h.render(w, r, "settings.html", map[string]any{
		"Roles":     roles,
		"UserRoles": roleIDs,
	})
}

func (h *Handler) handleMyProjects(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	projects, err := h.repo.ListUserProjects(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	type projectItem struct {
		Project       models.Project
		ResponseCount int
	}

	items := make([]projectItem, 0, len(projects))
	for _, p := range projects {
		count, _ := h.repo.CountProjectResponses(r.Context(), p.ID)
		items = append(items, projectItem{Project: p, ResponseCount: count})
	}

	h.render(w, r, "my_projects.html", map[string]any{
		"Items": items,
	})
}

func (h *Handler) handleMyResponses(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	responses, err := h.repo.ListUserResponses(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.render(w, r, "my_responses.html", map[string]any{
		"Responses": responses,
	})
}

func extractRoleIDs(roles []models.Role) []int64 {
	ids := make([]int64, len(roles))
	for i, r := range roles {
		ids[i] = r.ID
	}
	return ids
}
