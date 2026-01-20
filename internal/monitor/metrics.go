package monitor

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type FilterConfig struct {
	MountInclude  []string
	MountExclude  []string
	FstypeExclude []string
}

type DiskUsage struct {
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	UsedPercent float64 `json:"used_percent"`
	TotalBytes  uint64  `json:"total_bytes"`
	UsedBytes   uint64  `json:"used_bytes"`
}

type Metrics struct {
	Hostname   string      `json:"hostname"`
	CPUPercent float64     `json:"cpu_percent"`
	MemPercent float64     `json:"mem_percent"`
	Disks      []DiskUsage `json:"disks"`
}

func Collect(ctx context.Context, logger *zap.Logger, hostname string, filter FilterConfig) (Metrics, error) {
	cpuPercents, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, false)
	if err != nil {
		return Metrics{}, err
	}
	memStats, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return Metrics{}, err
	}
	partitions, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return Metrics{}, err
	}

	partitions = filterPartitions(partitions, filter)

	disks := make([]DiskUsage, 0, len(partitions))
	for _, part := range partitions {
		usage, err := disk.UsageWithContext(ctx, part.Mountpoint)
		if err != nil {
			logger.Debug("disk usage failed", zap.String("mountpoint", part.Mountpoint), zap.Error(err))
			continue
		}
		disks = append(disks, DiskUsage{
			Mountpoint:  part.Mountpoint,
			Fstype:      part.Fstype,
			UsedPercent: usage.UsedPercent,
			TotalBytes:  usage.Total,
			UsedBytes:   usage.Used,
		})
	}

	cpuPercent := 0.0
	if len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}

	return Metrics{
		Hostname:   hostname,
		CPUPercent: cpuPercent,
		MemPercent: memStats.UsedPercent,
		Disks:      disks,
	}, nil
}

func FormatMetricsHTML(metrics Metrics) string {
	var b strings.Builder
	host := html.EscapeString(metrics.Hostname)
	_, _ = fmt.Fprintf(&b, "<b>Simple System Monitor</b>\n<i>%s</i>\n\n", host)
	_, _ = fmt.Fprintf(&b, "<b>CPU</b>  %.1f%% %s\n", metrics.CPUPercent, statusEmoji(metrics.CPUPercent))
	_, _ = fmt.Fprintf(&b, "<b>MEM</b>  %.1f%% %s\n", metrics.MemPercent, statusEmoji(metrics.MemPercent))
	b.WriteString("\n<b>Disk</b>")
	if len(metrics.Disks) == 0 {
		b.WriteString("\n<pre>none</pre>")
		return b.String()
	}

	b.WriteString("\n<pre>")
	for i, d := range metrics.Disks {
		totalGiB := bytesToGiB(d.TotalBytes)
		usedGiB := bytesToGiB(d.UsedBytes)
		mount := html.EscapeString(d.Mountpoint)
		_, _ = fmt.Fprintf(&b, "%-16s %5.1f%% %s %5.1f/%-5.1fGiB", mount, d.UsedPercent, statusEmoji(d.UsedPercent), usedGiB, totalGiB)
		if i < len(metrics.Disks)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("</pre>")
	return b.String()
}

func bytesToGiB(value uint64) float64 {
	return float64(value) / (1024 * 1024 * 1024)
}

func statusEmoji(percent float64) string {
	switch {
	case percent >= 90:
		return "🟥"
	case percent >= 75:
		return "🟨"
	default:
		return "🟩"
	}
}
