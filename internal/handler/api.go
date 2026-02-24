package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"svyaz/internal/middleware"
)

func (h *Handler) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user := middleware.UserFromContext(r.Context())
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))

	if title == "" {
		http.Error(w, "Название обязательно", http.StatusBadRequest)
		return
	}

	stack := parseTags(r.FormValue("stack"))
	roleCounts := parseRoleCounts(r)

	projectID, err := h.repo.CreateProject(r.Context(), user.ID, title, description, stack, roleCounts)
	if err != nil {
		log.Printf("create project: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/project/%d?created=1", projectID), http.StatusFound)
}

func (h *Handler) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	stack := parseTags(r.FormValue("stack"))
	roleCounts := parseRoleCounts(r)

	if err := h.repo.UpdateProject(r.Context(), id, title, description, stack, roleCounts); err != nil {
		log.Printf("update project: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/project/%d", id), http.StatusFound)
}

func (h *Handler) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
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

	if err := h.repo.DeleteProject(r.Context(), id); err != nil {
		log.Printf("delete project: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/my/projects", http.StatusFound)
}

func (h *Handler) handleRespond(w http.ResponseWriter, r *http.Request) {
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
	if user.ID == project.AuthorID {
		http.Error(w, "Нельзя откликнуться на свой проект", http.StatusBadRequest)
		return
	}

	if already, _ := h.repo.HasUserResponded(r.Context(), id, user.ID); already {
		http.Redirect(w, r, fmt.Sprintf("/project/%d", id), http.StatusFound)
		return
	}

	if err := h.repo.CreateResponse(r.Context(), id, user.ID); err != nil {
		log.Printf("create response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	_ = h.repo.CreateNotification(r.Context(), project.AuthorID, "new_response", map[string]any{
		"project_id":    project.ID,
		"project_title": project.Title,
		"user_name":     user.Name,
		"user_id":       user.ID,
	})

	if h.tgClient != nil && project.Author != nil && project.Author.TgChatID > 0 {
		link := fmt.Sprintf("https://svyaz.fitra.tech/project/%d", project.ID)
		text := fmt.Sprintf("Новый отклик от <b>%s</b> на \"%s\"\n%s", user.Name, project.Title, link)
		go h.tgClient.SendMessage(project.Author.TgChatID, text)
	}

	http.Redirect(w, r, fmt.Sprintf("/project/%d", id), http.StatusFound)
}

func (h *Handler) handleCancelResponse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.UserFromContext(r.Context())

	resp, err := h.repo.GetUserResponseForProject(r.Context(), id, user.ID)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/project/%d", id), http.StatusFound)
		return
	}

	if resp.Status != "pending" {
		http.Redirect(w, r, fmt.Sprintf("/project/%d", id), http.StatusFound)
		return
	}

	if err := h.repo.DeleteResponse(r.Context(), resp.ID, user.ID); err != nil {
		log.Printf("cancel response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/project/%d", id), http.StatusFound)
}

func (h *Handler) handleUpdateResponse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	resp, err := h.repo.GetResponse(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	project, err := h.repo.GetProject(r.Context(), resp.ProjectID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user.ID != project.AuthorID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_ = r.ParseForm()
	status := r.FormValue("status")
	if status != "accepted" && status != "rejected" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateResponseStatus(r.Context(), id, status); err != nil {
		log.Printf("update response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if status == "accepted" {
		_ = h.repo.CreateNotification(r.Context(), resp.UserID, "response_accepted", map[string]any{
			"project_id":    project.ID,
			"project_title": project.Title,
		})

		if h.tgClient != nil {
			respUser, err := h.repo.GetUser(r.Context(), resp.UserID)
			if err == nil && respUser.TgChatID > 0 {
				link := fmt.Sprintf("https://svyaz.fitra.tech/project/%d", project.ID)
				text := fmt.Sprintf("Ваш отклик на \"%s\" принят!\n%s", project.Title, link)
				go h.tgClient.SendMessage(respUser.TgChatID, text)
			}
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/project/%d", project.ID), http.StatusFound)
}

func (h *Handler) handleSaveOnboarding(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user := middleware.UserFromContext(r.Context())
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = user.Name
	}

	bio := strings.TrimSpace(r.FormValue("bio"))
	experience := r.FormValue("experience")
	skills := parseTags(r.FormValue("skills"))
	roleIDs := parseIntSlice(r.Form["roles"])

	if err := h.repo.UpdateUserProfile(r.Context(), user.ID, name, bio, experience, skills, roleIDs); err != nil {
		log.Printf("update profile: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) handleSaveProfile(w http.ResponseWriter, r *http.Request) {
	h.handleSaveOnboarding(w, r)
}

func (h *Handler) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	notifs, err := h.repo.ListNotifications(r.Context(), user.ID, 20)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if notifs == nil {
		_, _ = w.Write([]byte("[]"))
		return
	}
	_ = json.NewEncoder(w).Encode(notifs)
}

func (h *Handler) handleMarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	_ = h.repo.MarkNotificationsRead(r.Context(), user.ID)
	w.WriteHeader(http.StatusOK)
}

func parseTags(s string) []string {
	var tags []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func parseIntSlice(ss []string) []int64 {
	var ids []int64
	for _, s := range ss {
		id, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func parseRoleCounts(r *http.Request) map[int64]int {
	rc := make(map[int64]int)
	for _, s := range r.Form["roles"] {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue
		}
		count := 1
		if v := r.FormValue(fmt.Sprintf("role_count_%d", id)); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 1 {
				count = n
			}
		}
		rc[id] = count
	}
	return rc
}
