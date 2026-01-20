package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	if err := client.SendMessage(context.Background(), "hello"); err != nil {
		t.Fatalf("expected send to succeed, got %v", err)
	}
	if !called {
		t.Fatalf("expected server to be called")
	}
}

func TestSendMessageNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewWithBaseURL("test", "chat", server.URL, server.Client())
	if err := client.SendMessage(context.Background(), "hello"); err == nil {
		t.Fatalf("expected error on non-2xx status")
	}
}

func TestSendMessageNilClient(t *testing.T) {
	var client *Client
	if err := client.SendMessage(context.Background(), "hello"); err == nil {
		t.Fatalf("expected error on nil client")
	}
}
