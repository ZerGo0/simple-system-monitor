package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type payloadCapture struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func TestNewRequiresCredentials(t *testing.T) {
	if New("", "") != nil {
		t.Fatalf("expected nil client with empty creds")
	}
	if New("token", "") != nil {
		t.Fatalf("expected nil client with empty chat id")
	}
}

func TestSendMessageUsesBaseURL(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/bottest/sendMessage" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewWithBaseURL("test", "chat", server.URL, server.Client())
	if err := client.SendHTMLMessage(context.Background(), "<b>hello</b>"); err != nil {
		t.Fatalf("expected send to succeed, got %v", err)
	}
	if !called {
		t.Fatalf("expected server to be called")
	}
}

func TestSendHTMLMessageSetsParseMode(t *testing.T) {
	var captured payloadCapture
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewWithBaseURL("test", "chat", server.URL, server.Client())
	if err := client.SendHTMLMessage(context.Background(), "<b>hi</b>"); err != nil {
		t.Fatalf("expected send to succeed, got %v", err)
	}
	if captured.ParseMode != "HTML" {
		t.Fatalf("expected parse_mode HTML, got %q", captured.ParseMode)
	}
}

func TestSendMessageNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewWithBaseURL("test", "chat", server.URL, server.Client())
	if err := client.SendHTMLMessage(context.Background(), "<b>hello</b>"); err == nil {
		t.Fatalf("expected error on non-2xx status")
	}
}

func TestSendMessageNilClient(t *testing.T) {
	var client *Client
	if err := client.SendHTMLMessage(context.Background(), "<b>hello</b>"); err == nil {
		t.Fatalf("expected error on nil client")
	}
}
