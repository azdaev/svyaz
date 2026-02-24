package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"svyaz/internal/config"
	"svyaz/internal/handler"
	"svyaz/internal/repo"
	"svyaz/internal/telegram"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := repo.New(cfg.DatabasePath, "migrations")
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	botUsername, err := fetchBotUsername(cfg.BotToken)
	if err != nil {
		log.Fatalf("telegram bot: %v", err)
	}
	log.Printf("Bot: @%s", botUsername)

	tgClient := telegram.NewClient(cfg.BotToken)

	ctx, cancel := context.WithCancel(context.Background())
	tgClient.StartPolling(ctx, func(tgUserID, chatID int64) {
		user, err := db.GetUserByTgID(context.Background(), tgUserID)
		if err != nil {
			log.Printf("poll: user not found for tg_id=%d: %v", tgUserID, err)
			return
		}
		if err := db.SetTgChatID(context.Background(), user.ID, chatID); err != nil {
			log.Printf("poll: set tg_chat_id: %v", err)
			return
		}
		log.Printf("poll: linked tg_chat_id=%d for user %d", chatID, user.ID)
		tgClient.SendMessage(chatID, "Уведомления подключены! Теперь вы будете получать сообщения о новых откликах.")
	})

	h := handler.New(db, "templates", cfg.BotToken, botUsername, cfg.CSRFSecret, cfg.CookieDomain, tgClient)
	router := h.Router()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
		log.Println("shutting down...")
		os.Exit(0)
	}()

	log.Printf("Starting server at http://%s", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), router); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func fetchBotUsername(token string) (string, error) {
	resp, err := http.Get("https://api.telegram.org/bot" + token + "/getMe")
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	if !result.OK || result.Result.Username == "" {
		return "", fmt.Errorf("invalid response from Telegram API")
	}

	return result.Result.Username, nil
}
