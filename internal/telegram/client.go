package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	token   string
	chatID  string
	client  *http.Client
	baseURL string
}

type message struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

const defaultBaseURL = "https://api.telegram.org"

func New(token string, chatID string) *Client {
	return NewWithBaseURL(token, chatID, defaultBaseURL, nil)
}

func NewWithBaseURL(token string, chatID string, baseURL string, httpClient *http.Client) *Client {
	if token == "" || chatID == "" {
		return nil
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		token:   token,
		chatID:  chatID,
		client:  httpClient,
		baseURL: baseURL,
	}
}

func (c *Client) SendMessage(ctx context.Context, text string) error {
	if c == nil {
		return errors.New("telegram client not configured")
	}
	payload := message{
		ChatID: c.chatID,
		Text:   text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", strings.TrimRight(c.baseURL, "/"), c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram API status %d", resp.StatusCode)
	}
	return nil
}
