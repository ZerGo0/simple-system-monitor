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
	DiskThreshold   float64
	DiskAlertWindow time.Duration
}

type AlertState struct {
	CPUAboveSince  time.Time
	CPUAlerting    bool
	MemAboveSince  time.Time
	MemAlerting    bool
	DiskAboveSince map[string]time.Time
	DiskAlerting   map[string]bool
}

func NewState() *AlertState {
	return &AlertState{
		DiskAboveSince: make(map[string]time.Time),
		DiskAlerting:   make(map[string]bool),
	}
}

func Check(metrics monitor.Metrics, cfg Thresholds, state *AlertState, now time.Time) []string {
	alerts := []string{}
	if cfg.CPUThreshold > 0 {
		if metrics.CPUPercent >= cfg.CPUThreshold {
			if state.CPUAboveSince.IsZero() {
				state.CPUAboveSince = now
			}
			if !state.CPUAlerting && now.Sub(state.CPUAboveSince) >= cfg.CPUAlertWindow {
				alerts = append(alerts, fmt.Sprintf("CPU %.1f%% >= %.1f%% for %s", metrics.CPUPercent, cfg.CPUThreshold, cfg.CPUAlertWindow))
				state.CPUAlerting = true
			}
		} else if state.CPUAlerting || !state.CPUAboveSince.IsZero() {
			state.CPUAlerting = false
			state.CPUAboveSince = time.Time{}
		}
	}
	if cfg.MemThreshold > 0 {
		if metrics.MemPercent >= cfg.MemThreshold {
			if state.MemAboveSince.IsZero() {
				state.MemAboveSince = now
			}
			if !state.MemAlerting && now.Sub(state.MemAboveSince) >= cfg.MemAlertWindow {
				alerts = append(alerts, fmt.Sprintf("Memory %.1f%% >= %.1f%% for %s", metrics.MemPercent, cfg.MemThreshold, cfg.MemAlertWindow))
				state.MemAlerting = true
			}
		} else if state.MemAlerting || !state.MemAboveSince.IsZero() {
			state.MemAlerting = false
			state.MemAboveSince = time.Time{}
		}
	}
	if cfg.DiskThreshold > 0 {
		for _, d := range metrics.Disks {
			mount := d.Mountpoint
			if d.UsedPercent >= cfg.DiskThreshold {
				if _, ok := state.DiskAboveSince[mount]; !ok {
					state.DiskAboveSince[mount] = now
				}
				if !state.DiskAlerting[mount] && now.Sub(state.DiskAboveSince[mount]) >= cfg.DiskAlertWindow {
					alerts = append(alerts, fmt.Sprintf("Disk %s %.1f%% >= %.1f%% for %s", mount, d.UsedPercent, cfg.DiskThreshold, cfg.DiskAlertWindow))
					state.DiskAlerting[mount] = true
				}
			} else {
				if state.DiskAlerting[mount] || !state.DiskAboveSince[mount].IsZero() {
					delete(state.DiskAlerting, mount)
					delete(state.DiskAboveSince, mount)
				}
			}
		}
		pruneDiskState(state, metrics.Disks)
	}
	return alerts
}

func pruneDiskState(state *AlertState, disks []monitor.DiskUsage) {
	active := make(map[string]struct{}, len(disks))
	for _, d := range disks {
		active[d.Mountpoint] = struct{}{}
	}
	for mount := range state.DiskAboveSince {
		if _, ok := active[mount]; !ok {
			delete(state.DiskAboveSince, mount)
		}
	}
	for mount := range state.DiskAlerting {
		if _, ok := active[mount]; !ok {
			delete(state.DiskAlerting, mount)
		}
	}
}
