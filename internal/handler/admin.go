package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.AdminStats(r.Context())
	if err != nil {
		log.Printf("admin stats: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.renderAdmin(w, r, "admin_dashboard.html", map[string]any{
		"Stats": stats,
	})
}

func (h *Handler) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	users, err := h.repo.AdminListUsers(r.Context(), search, 100, 0)
	if err != nil {
		log.Printf("admin list users: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.renderAdmin(w, r, "admin_users.html", map[string]any{
		"Users":  users,
		"Search": search,
	})
}

func (h *Handler) handleAdminProjects(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	projects, err := h.repo.AdminListProjects(r.Context(), search, status, 100, 0)
	if err != nil {
		log.Printf("admin list projects: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.renderAdmin(w, r, "admin_projects.html", map[string]any{
		"Projects":     projects,
		"Search":       search,
		"StatusFilter": status,
	})
}

func (h *Handler) handleAdminProjectView(w http.ResponseWriter, r *http.Request) {
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

	responses, _ := h.repo.ListProjectResponses(r.Context(), id)

	h.renderAdmin(w, r, "admin_project_view.html", map[string]any{
		"Project":   project,
		"Responses": responses,
	})
}

func (h *Handler) handleAdminToggleAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user, err := h.repo.GetUser(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.repo.SetUserAdmin(r.Context(), id, !user.IsAdmin); err != nil {
		log.Printf("toggle admin: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users", http.StatusFound)
}

func (h *Handler) handleAdminToggleBan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user, err := h.repo.GetUser(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.repo.SetUserBanned(r.Context(), id, !user.IsBanned); err != nil {
		log.Printf("toggle ban: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users", http.StatusFound)
}

func (h *Handler) handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.repo.AdminDeleteUser(r.Context(), id); err != nil {
		log.Printf("delete user: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users", http.StatusFound)
}

func (h *Handler) handleAdminApproveProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.repo.SetProjectStatus(r.Context(), id, "active"); err != nil {
		log.Printf("approve project: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects/"+chi.URLParam(r, "id"), http.StatusFound)
}

func (h *Handler) handleAdminHideProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.repo.SetProjectStatus(r.Context(), id, "hidden"); err != nil {
		log.Printf("hide project: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects/"+chi.URLParam(r, "id"), http.StatusFound)
}

func (h *Handler) handleAdminDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.repo.DeleteProject(r.Context(), id); err != nil {
		log.Printf("admin delete project: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/projects", http.StatusFound)
}
