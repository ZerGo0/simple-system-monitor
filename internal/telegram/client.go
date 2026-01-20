package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
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

func (c *Client) SendHTMLMessage(ctx context.Context, text string) error {
	return c.send(ctx, text, "HTML")
}

func (c *Client) SendPNG(ctx context.Context, filename string, data []byte) error {
	return c.SendPNGWithCaption(ctx, filename, data, "", "")
}

func (c *Client) SendPNGWithCaption(ctx context.Context, filename string, data []byte, caption string, parseMode string) error {
	if c == nil {
		return errors.New("telegram client not configured")
	}
	if filename == "" {
		filename = "message.png"
	}
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.WriteField("chat_id", c.chatID); err != nil {
		return err
	}
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return err
		}
		if parseMode != "" {
			if err := writer.WriteField("parse_mode", parseMode); err != nil {
				return err
			}
		}
	}
	part, err := writer.CreateFormFile("photo", filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/bot%s/sendPhoto", strings.TrimRight(c.baseURL, "/"), c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

func (c *Client) send(ctx context.Context, text string, parseMode string) error {
	if c == nil {
		return errors.New("telegram client not configured")
	}
	payload := message{
		ChatID:    c.chatID,
		Text:      text,
		ParseMode: parseMode,
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
