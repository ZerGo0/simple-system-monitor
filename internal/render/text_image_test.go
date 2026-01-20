package render

import (
	"bytes"
	"image/png"
	"testing"
)

func TestTextPNG(t *testing.T) {
	data, err := TextPNG("hello\nworld")
	if err != nil {
		t.Fatalf("expected render to succeed, got %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected png data")
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("expected valid png, got %v", err)
	}
}
