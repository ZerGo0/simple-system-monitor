package monitor

import "testing"

func TestBarWidth(t *testing.T) {
	barText := bar(50, 10)
	if barText != "[=====-----]" {
		t.Fatalf("unexpected bar: %s", barText)
	}
}

func TestStatusEmoji(t *testing.T) {
	if statusEmoji(10) != "🟩" {
		t.Fatalf("expected green for low usage")
	}
	if statusEmoji(80) != "🟨" {
		t.Fatalf("expected yellow for mid usage")
	}
	if statusEmoji(95) != "🟥" {
		t.Fatalf("expected red for high usage")
	}
}
