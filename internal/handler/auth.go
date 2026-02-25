package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"svyaz/internal/repo"
)

func (h *Handler) handleDevLogin(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Dev Login</title>
<style>
body{font-family:system-ui;max-width:480px;margin:40px auto;padding:0 20px;background:#f7f8fa}
h1{font-size:1.2rem;color:#1a1a2e;margin-bottom:24px}
a.user{display:flex;align-items:center;gap:12px;padding:12px 16px;background:#fff;border:1px solid #e5e7eb;border-radius:8px;margin-bottom:8px;text-decoration:none;color:#2d2d3f;transition:.15s}
a.user:hover{border-color:#5b9bd5;box-shadow:0 2px 8px rgba(91,155,213,.15)}
.avatar{width:36px;height:36px;border-radius:50%;background:#5b9bd5;color:#fff;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:.9rem;flex-shrink:0}
img.avatar{object-fit:cover}
.name{font-weight:600;font-size:.9rem}
.username{font-size:.8rem;color:#9ca3af}
</style></head><body><h1>Dev Login</h1>`)

	for _, u := range users {
		avatar := fmt.Sprintf(`<span class="avatar">%s</span>`, string([]rune(u.Name)[:1]))
		if u.PhotoURL != "" {
			avatar = fmt.Sprintf(`<img class="avatar" src="%s">`, u.PhotoURL)
		}
		fmt.Fprintf(w, `<a class="user" href="/auth/dev/%d">%s<div><div class="name">%s</div><div class="username">@%s</div></div></a>`,
			u.ID, avatar, u.Name, u.TgUsername)
	}

	fmt.Fprint(w, `</body></html>`)
}

func (h *Handler) handleDevLoginAs(w http.ResponseWriter, r *http.Request) {
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

	token := repo.GenerateToken()
	if err := h.repo.CreateSession(r.Context(), token, user.ID); err != nil {
		log.Printf("dev login session: %v", err)
		http.Error(w, "Internal error", 500)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   30 * 24 * 3600,
		SameSite: http.SameSiteLaxMode,
	})

	if !user.Onboarded {
		http.Redirect(w, r, "/onboarding", http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (h *Handler) handleTelegramAuth(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if !h.validateTelegramAuth(query) {
		http.Error(w, "Ошибка авторизации", http.StatusForbidden)
		return
	}

	tgID, err := strconv.ParseInt(query.Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	username := query.Get("username")
	name := query.Get("first_name")
	if ln := query.Get("last_name"); ln != "" {
		name += " " + ln
	}
	photoURL := query.Get("photo_url")

	user, isNew, err := h.repo.UpsertUser(r.Context(), tgID, username, name, photoURL)
	if err != nil {
		log.Printf("upsert user error: %v", err)
		http.Error(w, "Internal error", 500)
		return
	}

	token := repo.GenerateToken()
	if err := h.repo.CreateSession(r.Context(), token, user.ID); err != nil {
		log.Printf("create session error: %v", err)
		http.Error(w, "Internal error", 500)
		return
	}

	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	// Clear old host-only cookie (no Domain) to avoid conflicts with new domain cookie
	if h.cookieDomain != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})
	}

	cookie := &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		MaxAge:   30 * 24 * 3600,
		SameSite: http.SameSiteLaxMode,
	}
	if h.cookieDomain != "" {
		cookie.Domain = h.cookieDomain
	}
	http.SetCookie(w, cookie)

	if isNew || !user.Onboarded {
		http.Redirect(w, r, "/onboarding", http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		_ = h.repo.DeleteSession(r.Context(), cookie.Value)
	}

	// Clear host-only cookie (no Domain)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	// Clear domain cookie
	if h.cookieDomain != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			Domain:   h.cookieDomain,
			HttpOnly: true,
			MaxAge:   -1,
		})
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) validateTelegramAuth(query url.Values) bool {
	hash := query.Get("hash")
	if hash == "" {
		return false
	}

	params := make([]string, 0, len(query))
	for key, values := range query {
		if key == "hash" {
			continue
		}
		params = append(params, key+"="+values[0])
	}
	sort.Strings(params)

	dataCheckString := ""
	for i, p := range params {
		if i > 0 {
			dataCheckString += "\n"
		}
		dataCheckString += p
	}

	secretKey := sha256.Sum256([]byte(h.botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(hash), []byte(expected)) {
		return false
	}

	// Reject auth data older than 1 day
	if authDate, err := strconv.ParseInt(query.Get("auth_date"), 10, 64); err == nil {
		age := math.Abs(float64(time.Now().Unix() - authDate))
		if age > 86400 {
			return false
		}
	}

	return true
}
