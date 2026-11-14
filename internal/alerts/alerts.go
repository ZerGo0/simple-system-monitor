package alerts

import (
	"fmt"
	"time"

	"github.com/zergo0/simple-system-monitor/internal/monitor"
)

type Thresholds struct {
	CPUThreshold    float64
	CPUAlertWindow  time.Duration
	MemThreshold    float64
	MemAlertWindow  time.Duration
	SwapThreshold   float64
	SwapAlertWindow time.Duration
	DiskThreshold   float64
	DiskAlertWindow time.Duration
}

type windowState struct {
	aboveSince time.Time
	alerting   bool
}

func (w *windowState) check(value, threshold float64, window time.Duration, now time.Time) bool {
	if threshold <= 0 || value < threshold {
		*w = windowState{}
		return false
	}
	if w.aboveSince.IsZero() {
		w.aboveSince = now
	}
	if w.alerting || now.Sub(w.aboveSince) < window {
		return false
	}
	w.alerting = true
	return true
}

type AlertState struct {
	cpu  windowState
	mem  windowState
	swap windowState
	disk map[string]*windowState
}

func NewState() *AlertState {
	return &AlertState{disk: make(map[string]*windowState)}
}

func Check(metrics monitor.Metrics, cfg Thresholds, state *AlertState, now time.Time) []string {
	alerts := []string{}
	if state.cpu.check(metrics.CPUPercent, cfg.CPUThreshold, cfg.CPUAlertWindow, now) {
		alerts = append(alerts, fmt.Sprintf("CPU %.1f%% >= %.1f%% for %s", metrics.CPUPercent, cfg.CPUThreshold, cfg.CPUAlertWindow))
	}
	if state.mem.check(metrics.MemPercent, cfg.MemThreshold, cfg.MemAlertWindow, now) {
		alerts = append(alerts, fmt.Sprintf("Memory %.1f%% >= %.1f%% for %s", metrics.MemPercent, cfg.MemThreshold, cfg.MemAlertWindow))
	}
	if state.swap.check(metrics.SwapPercent, cfg.SwapThreshold, cfg.SwapAlertWindow, now) {
		alerts = append(alerts, fmt.Sprintf("Swap %.1f%% >= %.1f%% for %s", metrics.SwapPercent, cfg.SwapThreshold, cfg.SwapAlertWindow))
	}
	active := make(map[string]struct{}, len(metrics.Disks))
	for _, d := range metrics.Disks {
		active[d.Mountpoint] = struct{}{}
		w, ok := state.disk[d.Mountpoint]
		if !ok {
			w = &windowState{}
			state.disk[d.Mountpoint] = w
		}
		if w.check(d.UsedPercent, cfg.DiskThreshold, cfg.DiskAlertWindow, now) {
			alerts = append(alerts, fmt.Sprintf("Disk %s %.1f%% >= %.1f%% for %s", d.Mountpoint, d.UsedPercent, cfg.DiskThreshold, cfg.DiskAlertWindow))
		}
	}
	for mount := range state.disk {
		if _, ok := active[mount]; !ok {
			delete(state.disk, mount)
		}
	}
	return alerts
}
