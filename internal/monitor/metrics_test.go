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

func TestTableRowWidth(t *testing.T) {
	row := tableRow(4, 4, 2, 6, []string{"mnt", "1.0%", "🟩", "1/2GiB"})
	if len(row) == 0 || !strings.HasPrefix(row, "│") {
		t.Fatalf("expected table row formatting")
	}
}

func TestTableRow3Width(t *testing.T) {
	row := tableRow3(5, 4, 2, []string{"CPU", "1.0%", "🟩"})
	if len(row) == 0 || !strings.HasPrefix(row, "│") {
		t.Fatalf("expected table row formatting")
	}
}

func TestBuildMetricRowsUsedTotal(t *testing.T) {
	metrics := Metrics{CPUPercent: 25, CPUCores: 8, MemUsedBytes: 4 << 30, MemTotalBytes: 16 << 30, SwapUsedBytes: 1 << 29, SwapTotalBytes: 2 << 30}
	rows := buildMetricRows(metrics, statusLabel)
	want := []string{"2.0/8 cores", "4.0/16.0GiB", "0.5/2.0GiB"}
	if len(rows) != len(want) {
		t.Fatalf("expected %d rows, got %d", len(want), len(rows))
	}
	for i, row := range rows {
		if row[3] != want[i] {
			t.Fatalf("row %s: expected %q, got %q", row[0], want[i], row[3])
		}
	}
}
