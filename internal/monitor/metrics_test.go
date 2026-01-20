package monitor

import (
	"strings"
	"testing"
)

func TestFormatMount(t *testing.T) {
	if got := formatMount("/short", 10); got != "/short" {
		t.Fatalf("expected mount unchanged, got %q", got)
	}
	if got := formatMount("/very/long/mountpoint", 8); got != "/very/l…" {
		t.Fatalf("unexpected truncation: %q", got)
	}
}

func TestTableRow2Width(t *testing.T) {
	row := tableRow2(5, 6, []string{"Mount", "1/2GiB"})
	if len(row) == 0 || !strings.HasPrefix(row, "│") {
		t.Fatalf("expected table row formatting")
	}
}

func TestTableRow3Width(t *testing.T) {
	row := tableRow3(5, 4, 2, []string{"CPU", "1.0%", "OK"})
	if len(row) == 0 || !strings.HasPrefix(row, "│") {
		t.Fatalf("expected table row formatting")
	}
}
