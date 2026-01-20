package telegram

import (
	"bytes"
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

func TestSendPNGUsesBaseURL(t *testing.T) {
	called := false
	var gotChatID string
	var gotFilename string
	var gotPayload []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/bottest/sendPhoto" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gotChatID = r.FormValue("chat_id")
		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		gotFilename = header.Filename
		gotPayload, _ = io.ReadAll(file)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewWithBaseURL("test", "chat", server.URL, server.Client())
	payload := []byte{0x89, 0x50, 0x4e, 0x47}
	if err := client.SendPNG(context.Background(), "metrics.png", payload); err != nil {
		t.Fatalf("expected send to succeed, got %v", err)
	}
	if !called {
		t.Fatalf("expected server to be called")
	}
	if gotChatID != "chat" {
		t.Fatalf("expected chat id to be set, got %q", gotChatID)
	}
	if gotFilename != "metrics.png" {
		t.Fatalf("expected filename metrics.png, got %q", gotFilename)
	}
	if !bytes.Equal(gotPayload, payload) {
		t.Fatalf("expected payload to match")
	}
}

func TestSendPNGWithCaption(t *testing.T) {
	var gotChatID string
	var gotCaption string
	var gotParseMode string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottest/sendPhoto" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gotChatID = r.FormValue("chat_id")
		gotCaption = r.FormValue("caption")
		gotParseMode = r.FormValue("parse_mode")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewWithBaseURL("test", "chat", server.URL, server.Client())
	if err := client.SendPNGWithCaption(context.Background(), "metrics.png", []byte{0x01}, "<b>Hi</b>", "HTML"); err != nil {
		t.Fatalf("expected send to succeed, got %v", err)
	}
	if gotChatID != "chat" {
		t.Fatalf("expected chat id to be set, got %q", gotChatID)
	}
	if gotCaption != "<b>Hi</b>" {
		t.Fatalf("expected caption to be set, got %q", gotCaption)
	}
	if gotParseMode != "HTML" {
		t.Fatalf("expected parse mode HTML, got %q", gotParseMode)
	}
}

func TestSendPNGNilClient(t *testing.T) {
	var client *Client
	if err := client.SendPNG(context.Background(), "message.png", []byte{0x00}); err == nil {
		t.Fatalf("expected error on nil client")
	}
}
