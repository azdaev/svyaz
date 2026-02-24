package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	token string
	http  *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		http:  &http.Client{Timeout: 35 * time.Second},
	}
}

func (c *Client) SendMessage(chatID int64, text string) {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.token)

	resp, err := c.http.PostForm(endpoint, url.Values{
		"chat_id":                  {fmt.Sprintf("%d", chatID)},
		"text":                     {text},
		"parse_mode":               {"HTML"},
		"disable_web_page_preview": {"true"},
	})
	if err != nil {
		log.Printf("telegram send: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var result struct {
			Description string `json:"description"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&result)
		log.Printf("telegram send error: %s", result.Description)
	}
}

type update struct {
	UpdateID int64 `json:"update_id"`
	Message  *message `json:"message"`
}

type message struct {
	Text string  `json:"text"`
	From *tgUser `json:"from"`
	Chat *tgChat `json:"chat"`
}

type tgUser struct {
	ID int64 `json:"id"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

// StartHandler is called for each /start message with (tgUserID, chatID).
type StartHandler func(tgUserID, chatID int64)

// StartPolling runs long polling for /start commands in a background goroutine.
// It clears any existing webhook and polls getUpdates.
func (c *Client) StartPolling(ctx context.Context, onStart StartHandler) {
	// Clear webhook so polling works
	c.deleteWebhook()

	go func() {
		offset := int64(0)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			updates, err := c.getUpdates(offset, 30)
			if err != nil {
				log.Printf("telegram poll: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, u := range updates {
				offset = u.UpdateID + 1
				if u.Message != nil && u.Message.Text == "/start" && u.Message.From != nil && u.Message.Chat != nil {
					onStart(u.Message.From.ID, u.Message.Chat.ID)
				}
			}
		}
	}()
}

func (c *Client) getUpdates(offset int64, timeout int) ([]update, error) {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=%d&allowed_updates=[\"message\"]",
		c.token, offset, timeout)

	resp, err := c.http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool     `json:"ok"`
		Result []update `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("getUpdates: not ok")
	}
	return result.Result, nil
}

func (c *Client) deleteWebhook() {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook", c.token)
	resp, err := c.http.Get(endpoint)
	if err != nil {
		log.Printf("telegram deleteWebhook: %v", err)
		return
	}
	resp.Body.Close()
}
