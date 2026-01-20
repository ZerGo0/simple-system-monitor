package alerts

import (
	"testing"
	"time"

	"simple-system-monitor/internal/monitor"
)

func TestCPUAlertWindow(t *testing.T) {
	state := NewState()
	cfg := Thresholds{CPUThreshold: 80, CPUAlertWindow: 5 * time.Minute}
	metrics := monitor.Metrics{CPUPercent: 90}

	start := time.Now()
	alerts := Check(metrics, cfg, state, start)
	if len(alerts) != 0 {
		t.Fatalf("expected no alert before window")
	}

	alerts = Check(metrics, cfg, state, start.Add(5*time.Minute))
	if len(alerts) != 1 {
		t.Fatalf("expected alert after window")
	}

	alerts = Check(metrics, cfg, state, start.Add(6*time.Minute))
	if len(alerts) != 0 {
		t.Fatalf("expected no repeat alert")
	}

	metrics.CPUPercent = 10
	alerts = Check(metrics, cfg, state, start.Add(7*time.Minute))
	if len(alerts) != 0 {
		t.Fatalf("expected no alert after reset")
	}
}

func TestMemAlertWindow(t *testing.T) {
	state := NewState()
	cfg := Thresholds{MemThreshold: 70, MemAlertWindow: 2 * time.Minute}
	metrics := monitor.Metrics{MemPercent: 80}

	start := time.Now()
	alerts := Check(metrics, cfg, state, start)
	if len(alerts) != 0 {
		t.Fatalf("expected no alert before window")
	}

	alerts = Check(metrics, cfg, state, start.Add(2*time.Minute))
	if len(alerts) != 1 {
		t.Fatalf("expected alert after window")
	}
}

func TestDiskAlertWindow(t *testing.T) {
	state := NewState()
	cfg := Thresholds{DiskThreshold: 80, DiskAlertWindow: 3 * time.Minute}
	metrics := monitor.Metrics{Disks: []monitor.DiskUsage{{Mountpoint: "/", UsedPercent: 85}}}

	start := time.Now()
	alerts := Check(metrics, cfg, state, start)
	if len(alerts) != 0 {
		t.Fatalf("expected no alert before window")
	}

	alerts = Check(metrics, cfg, state, start.Add(3*time.Minute))
	if len(alerts) != 1 {
		t.Fatalf("expected alert after window")
	}

	metrics.Disks = []monitor.DiskUsage{{Mountpoint: "/", UsedPercent: 10}}
	alerts = Check(metrics, cfg, state, start.Add(4*time.Minute))
	if len(alerts) != 0 {
		t.Fatalf("expected no alert after reset")
	}
}
