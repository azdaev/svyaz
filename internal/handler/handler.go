package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"svyaz/internal/middleware"
	"svyaz/internal/repo"
)

type Handler struct {
	repo         *repo.Repo
	tmplDir      string
	botToken     string
	botUsername  string
	csrfSecret   string
	cookieDomain string
}

func New(r *repo.Repo, tmplDir, botToken, botUsername, csrfSecret, cookieDomain string) *Handler {
	return &Handler{
		repo:         r,
		tmplDir:      tmplDir,
		botToken:     botToken,
		botUsername:  botUsername,
		csrfSecret:   csrfSecret,
		cookieDomain: cookieDomain,
	}
}

func (h *Handler) Router() http.Handler {
	main := h.mainRouter()
	admin := h.adminRouter()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Host, "admin.") {
			admin.ServeHTTP(w, r)
		} else {
			main.ServeHTTP(w, r)
		}
	})
}

func (h *Handler) mainRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.CleanPath)
	r.Use(middleware.Auth(h.repo))

	// Static files
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	r.Handle("/static/*", fs)

	// Pages
	r.Get("/", h.handleIndex)
	r.Get("/project/new", h.requireAuth(h.handleProjectNew))
	r.Get("/project/{id}", h.handleProjectView)
	r.Get("/project/{id}/edit", h.requireAuth(h.handleProjectEdit))
	r.Get("/user/{id}", h.handleUserProfile)
	r.Get("/onboarding", h.requireAuth(h.handleOnboarding))
	r.Get("/settings", h.requireAuth(h.handleSettings))
	r.Get("/my/projects", h.requireAuth(h.handleMyProjects))
	r.Get("/my/responses", h.requireAuth(h.handleMyResponses))

	// Auth
	r.Get("/auth/telegram", h.handleTelegramAuth)
	r.Post("/auth/logout", h.handleLogout)

	// API
	r.Route("/api", func(r chi.Router) {
		r.Use(h.csrfMiddleware)

		r.Post("/projects", h.requireAuth(h.handleCreateProject))
		r.Post("/projects/{id}", h.requireAuth(h.handleUpdateProject))
		r.Post("/projects/{id}/delete", h.requireAuth(h.handleDeleteProject))
		r.Post("/projects/{id}/respond", h.requireAuth(h.handleRespond))
		r.Post("/responses/{id}", h.requireAuth(h.handleUpdateResponse))
		r.Post("/user/onboarding", h.requireAuth(h.handleSaveOnboarding))
		r.Post("/user/profile", h.requireAuth(h.handleSaveProfile))
		r.Get("/notifications", h.requireAuth(h.handleGetNotifications))
		r.Post("/notifications/read", h.requireAuth(h.handleMarkNotificationsRead))
	})

	return r
}

func (h *Handler) adminRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.CleanPath)
	r.Use(middleware.Auth(h.repo))

	fs := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	r.Handle("/static/*", fs)

	r.Group(func(r chi.Router) {
		r.Use(h.requireAdmin)

		r.Get("/", h.handleAdminDashboard)
		r.Get("/users", h.handleAdminUsers)
		r.Get("/projects", h.handleAdminProjects)
		r.Get("/projects/{id}", h.handleAdminProjectView)

		r.Route("/api", func(r chi.Router) {
			r.Use(h.csrfMiddleware)

			r.Post("/users/{id}/toggle-admin", h.handleAdminToggleAdmin)
			r.Post("/users/{id}/toggle-ban", h.handleAdminToggleBan)
			r.Post("/users/{id}/delete", h.handleAdminDeleteUser)
			r.Post("/projects/{id}/approve", h.handleAdminApproveProject)
			r.Post("/projects/{id}/hide", h.handleAdminHideProject)
			r.Post("/projects/{id}/delete", h.handleAdminDeleteProject)
		})
	})

	// Auth (accessible without admin check)
	r.Post("/auth/logout", h.handleLogout)

	return r
}

func (h *Handler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := middleware.UserFromContext(r.Context())
		if user == nil || !user.IsAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if middleware.UserFromContext(r.Context()) == nil {
			http.Redirect(w, r, "/?login=1", http.StatusFound)
			return
		}
		next(w, r)
	}
}

func (h *Handler) csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}
		if !h.validateCSRF(r) {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, page string, data map[string]any) {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			months := []string{
				"", "янв", "фев", "мар", "апр", "мая", "июн",
				"июл", "авг", "сен", "окт", "ноя", "дек",
			}
			return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()], t.Year())
		},
		"truncate": func(s string, n int) string {
			runes := []rune(s)
			if len(runes) <= n {
				return s
			}
			return string(runes[:n]) + "..."
		},
		"join": strings.Join,
		"statusText": func(s string) string {
			m := map[string]string{
				"pending":  "На рассмотрении",
				"accepted": "Принят",
				"rejected": "Отклонён",
			}
			if v, ok := m[s]; ok {
				return v
			}
			return s
		},
		"statusClass": func(s string) string { return s },
		"plural": func(n int, one, few, many string) string {
			if n%10 == 1 && n%100 != 11 {
				return fmt.Sprintf("%d %s", n, one)
			}
			if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
				return fmt.Sprintf("%d %s", n, few)
			}
			return fmt.Sprintf("%d %s", n, many)
		},
		"hasRole": func(roles []int64, id int64) bool {
			for _, r := range roles {
				if r == id {
					return true
				}
			}
			return false
		},
		"slice": func(s string, start, end int) string {
			runes := []rune(s)
			if start >= len(runes) {
				return ""
			}
			if end > len(runes) {
				end = len(runes)
			}
			return string(runes[start:end])
		},
		"roleCount": func(m map[int64]int, id int64) int {
			if c, ok := m[id]; ok {
				return c
			}
			return 1
		},
	}

	if data == nil {
		data = make(map[string]any)
	}

	user := middleware.UserFromContext(r.Context())
	data["User"] = user
	data["BotUsername"] = h.botUsername

	if user != nil {
		data["CSRFToken"] = h.generateCSRF(r)
		count, _ := h.repo.UnreadNotificationCount(r.Context(), user.ID)
		data["NotifCount"] = count
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join(h.tmplDir, "base.html"),
		filepath.Join(h.tmplDir, page),
	)
	if err != nil {
		log.Printf("template parse error (%s): %v", page, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("template execute error (%s): %v", page, err)
	}
}

func (h *Handler) renderAdmin(w http.ResponseWriter, r *http.Request, page string, data map[string]any) {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			months := []string{
				"", "янв", "фев", "мар", "апр", "мая", "июн",
				"июл", "авг", "сен", "окт", "ноя", "дек",
			}
			return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()], t.Year())
		},
		"join": strings.Join,
		"truncate": func(s string, n int) string {
			runes := []rune(s)
			if len(runes) <= n {
				return s
			}
			return string(runes[:n]) + "..."
		},
		"slice": func(s string, start, end int) string {
			runes := []rune(s)
			if start >= len(runes) {
				return ""
			}
			if end > len(runes) {
				end = len(runes)
			}
			return string(runes[start:end])
		},
	}

	if data == nil {
		data = make(map[string]any)
	}

	user := middleware.UserFromContext(r.Context())
	data["User"] = user
	if user != nil {
		data["CSRFToken"] = h.generateCSRF(r)
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join(h.tmplDir, "admin_base.html"),
		filepath.Join(h.tmplDir, page),
	)
	if err != nil {
		log.Printf("admin template parse error (%s): %v", page, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("admin template execute error (%s): %v", page, err)
	}
}

func (h *Handler) generateCSRF(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(h.csrfSecret))
	mac.Write([]byte(cookie.Value))
	return hex.EncodeToString(mac.Sum(nil))[:32]
}

func (h *Handler) validateCSRF(r *http.Request) bool {
	expected := h.generateCSRF(r)
	if expected == "" {
		return false
	}
	_ = r.ParseForm()
	token := r.FormValue("csrf_token")
	if token == "" {
		token = r.Header.Get("X-CSRF-Token")
	}
	return hmac.Equal([]byte(token), []byte(expected))
}
