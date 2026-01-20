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

	maxMount := maxMountWidth(metrics.Disks, 24)
	const usageWidth = 5
	const statusWidth = 2
	const sizeWidth = 14
	header := []string{"Mount", "Use", "St", "Used/Total"}
	b.WriteString("\n<pre>")
	b.WriteString(tableTop(maxMount, usageWidth, statusWidth, sizeWidth))
	b.WriteString(tableRow(maxMount, usageWidth, statusWidth, sizeWidth, header))
	b.WriteString(tableMid(maxMount, usageWidth, statusWidth, sizeWidth))
	for i, d := range metrics.Disks {
		totalGiB := bytesToGiB(d.TotalBytes)
		usedGiB := bytesToGiB(d.UsedBytes)
		mount := html.EscapeString(formatMount(d.Mountpoint, maxMount))
		use := fmt.Sprintf("%.1f%%", d.UsedPercent)
		status := statusEmoji(d.UsedPercent)
		size := fmt.Sprintf("%.1f/%.1fGiB", usedGiB, totalGiB)
		b.WriteString(tableRow(maxMount, usageWidth, statusWidth, sizeWidth, []string{mount, use, status, size}))
		if i < len(metrics.Disks)-1 {
			b.WriteString(tableMid(maxMount, usageWidth, statusWidth, sizeWidth))
		}
	}
	b.WriteString(tableBottom(maxMount, usageWidth, statusWidth, sizeWidth))
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

func maxMountWidth(disks []DiskUsage, max int) int {
	width := 1
	for _, d := range disks {
		if len(d.Mountpoint) > width {
			width = len(d.Mountpoint)
		}
	}
	if max > 0 && width > max {
		return max
	}
	return width
}

func formatMount(mount string, width int) string {
	if width <= 0 || len(mount) <= width {
		return mount
	}
	if width <= 3 {
		return mount[:width]
	}
	return mount[:width-1] + "…"
}

func tableTop(mountW, useW, statusW, sizeW int) string {
	return fmt.Sprintf("┌%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", mountW+2),
		strings.Repeat("─", useW+2),
		strings.Repeat("─", statusW+2),
		strings.Repeat("─", sizeW+2),
	)
}

func tableMid(mountW, useW, statusW, sizeW int) string {
	return fmt.Sprintf("├%s┼%s┼%s┼%s┤\n",
		strings.Repeat("─", mountW+2),
		strings.Repeat("─", useW+2),
		strings.Repeat("─", statusW+2),
		strings.Repeat("─", sizeW+2),
	)
}

func tableBottom(mountW, useW, statusW, sizeW int) string {
	return fmt.Sprintf("└%s┴%s┴%s┴%s┘",
		strings.Repeat("─", mountW+2),
		strings.Repeat("─", useW+2),
		strings.Repeat("─", statusW+2),
		strings.Repeat("─", sizeW+2),
	)
}

func tableRow(mountW, useW, statusW, sizeW int, cols []string) string {
	mount := padRight(cols[0], mountW)
	use := padLeft(cols[1], useW)
	status := padRight(cols[2], statusW)
	size := padRight(cols[3], sizeW)
	return fmt.Sprintf("│ %s │ %s │ %s │ %s │\n", mount, use, status, size)
}

func padRight(value string, width int) string {
	if width <= 0 {
		return value
	}
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}

func padLeft(value string, width int) string {
	if width <= 0 {
		return value
	}
	if len(value) >= width {
		return value
	}
	return strings.Repeat(" ", width-len(value)) + value
}
