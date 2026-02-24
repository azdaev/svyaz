package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"svyaz/internal/repo"
)

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

	user, isNew, err := h.repo.UpsertUser(r.Context(), tgID, username, name)
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

	logoutCookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	if h.cookieDomain != "" {
		logoutCookie.Domain = h.cookieDomain
	}
	http.SetCookie(w, logoutCookie)

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
