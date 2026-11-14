package monitor

import "testing"

func TestFormatMountPlain(t *testing.T) {
	if got := formatMountPlain("/short", 10); got != "/short" {
		t.Fatalf("expected mount unchanged, got %q", got)
	}
	if got := formatMountPlain("/very/long/mountpoint", 8); got != "/very..." {
		t.Fatalf("unexpected truncation: %q", got)
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
