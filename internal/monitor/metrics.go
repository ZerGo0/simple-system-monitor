package monitor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

const initialCPUSampleInterval = time.Second

var cpuBaseline sync.Once

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
	Hostname       string      `json:"hostname"`
	CPUPercent     float64     `json:"cpu_percent"`
	CPUCores       int         `json:"cpu_cores"`
	MemPercent     float64     `json:"mem_percent"`
	MemUsedBytes   uint64      `json:"mem_used_bytes"`
	MemTotalBytes  uint64      `json:"mem_total_bytes"`
	SwapPercent    float64     `json:"swap_percent"`
	SwapUsedBytes  uint64      `json:"swap_used_bytes"`
	SwapTotalBytes uint64      `json:"swap_total_bytes"`
	Disks          []DiskUsage `json:"disks"`
}

func Collect(ctx context.Context, logger *zap.Logger, hostname string, filter FilterConfig) (Metrics, error) {
	cpuPercent, err := collectCPUPercent(ctx)
	if err != nil {
		return Metrics{}, err
	}
	cpuCores, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		logger.Debug("cpu count failed", zap.Error(err))
	}
	memStats, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return Metrics{}, err
	}
	swapStats, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		logger.Debug("swap usage failed", zap.Error(err))
		swapStats = &mem.SwapMemoryStat{}
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

	return Metrics{
		Hostname:       hostname,
		CPUPercent:     cpuPercent,
		CPUCores:       cpuCores,
		MemPercent:     memStats.UsedPercent,
		MemUsedBytes:   memStats.Used,
		MemTotalBytes:  memStats.Total,
		SwapPercent:    swapStats.UsedPercent,
		SwapUsedBytes:  swapStats.Used,
		SwapTotalBytes: swapStats.Total,
		Disks:          disks,
	}, nil
}

func collectCPUPercent(ctx context.Context) (float64, error) {
	interval := time.Duration(0)
	cpuBaseline.Do(func() {
		interval = initialCPUSampleInterval
	})
	percents, err := cpu.PercentWithContext(ctx, interval, false)
	if err != nil || len(percents) == 0 {
		return 0, err
	}
	return percents[0], nil
}

func FormatMetricsHeaderText(metrics Metrics) string {
	host := CleanText(metrics.Hostname)
	if host == "" {
		return "Simple System Monitor"
	}
	return fmt.Sprintf("Simple System Monitor - %s", host)
}

func FormatMetricsText(metrics Metrics) string {
	lines := []string{}

	metricHeader := []string{"Metric", "Usage", "Status", "Used/Total"}
	metricRows := buildMetricRows(metrics, statusLabel)
	lines = append(lines, formatTableLines(metricHeader, metricRows, []bool{false, true, false, true})...)
	lines = append(lines, "", "Disk")
	if len(metrics.Disks) == 0 {
		lines = append(lines, "none")
		return strings.Join(lines, "\n")
	}

	maxMount := maxMountWidth(metrics.Disks, 24)
	diskHeader := []string{"Mount", "Usage", "Status", "Used/Total"}
	diskRows := make([][]string, 0, len(metrics.Disks))
	for _, d := range metrics.Disks {
		mount := formatMountPlain(CleanText(d.Mountpoint), maxMount)
		use := fmt.Sprintf("%.1f%%", d.UsedPercent)
		status := statusLabel(d.UsedPercent)
		size := formatUsedTotalGiB(d.UsedBytes, d.TotalBytes)
		diskRows = append(diskRows, []string{mount, use, status, size})
	}
	lines = append(lines, formatTableLines(diskHeader, diskRows, []bool{false, true, false, true})...)
	return strings.Join(lines, "\n")
}

func buildMetricRows(metrics Metrics, status func(float64) string) [][]string {
	rows := [][]string{
		{"CPU", fmt.Sprintf("%.1f%%", metrics.CPUPercent), status(metrics.CPUPercent), formatUsedTotalCores(metrics.CPUPercent, metrics.CPUCores)},
		{"MEM", fmt.Sprintf("%.1f%%", metrics.MemPercent), status(metrics.MemPercent), formatUsedTotalGiB(metrics.MemUsedBytes, metrics.MemTotalBytes)},
	}
	if metrics.SwapTotalBytes > 0 {
		rows = append(rows, []string{"SWAP", fmt.Sprintf("%.1f%%", metrics.SwapPercent), status(metrics.SwapPercent), formatUsedTotalGiB(metrics.SwapUsedBytes, metrics.SwapTotalBytes)})
	}
	return rows
}

func formatUsedTotalCores(percent float64, cores int) string {
	if cores <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f/%d cores", percent/100*float64(cores), cores)
}

func formatUsedTotalGiB(used, total uint64) string {
	return fmt.Sprintf("%.1f/%.1fGiB", bytesToGiB(used), bytesToGiB(total))
}

func bytesToGiB(value uint64) float64 {
	return float64(value) / (1024 * 1024 * 1024)
}

func statusLabel(percent float64) string {
	switch {
	case percent >= 90:
		return "ALERT"
	case percent >= 75:
		return "WARN"
	default:
		return "OK"
	}
}

func maxMountWidth(disks []DiskUsage, max int) int {
	width := 1
	for _, d := range disks {
		if runeLen(d.Mountpoint) > width {
			width = runeLen(d.Mountpoint)
		}
	}
	if max > 0 && width > max {
		return max
	}
	return width
}

func formatMountPlain(mount string, width int) string {
	return formatMountEllipsis(mount, width, "...")
}

func formatMountEllipsis(mount string, width int, ellipsis string) string {
	if width <= 0 || runeLen(mount) <= width {
		return mount
	}
	ellipsisLen := runeLen(ellipsis)
	if width <= ellipsisLen {
		return string([]rune(mount)[:width])
	}
	return string([]rune(mount)[:width-ellipsisLen]) + ellipsis
}

func runeLen(value string) int {
	return utf8.RuneCountInString(value)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func displayWidth(value string) int {
	width := 0
	for _, r := range value {
		width += runeDisplayWidth(r)
	}
	return width
}

func CleanText(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	return value
}

func runeDisplayWidth(r rune) int {
	switch {
	case r >= 0x1F000:
		return 2
	case r >= 0x2600 && r <= 0x27BF:
		return 2
	case r == '…':
		return 1
	default:
		return 1
	}
}

func formatTableLines(header []string, rows [][]string, rightAlign []bool) []string {
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = displayWidth(h)
	}
	for _, row := range rows {
		for i, col := range row {
			widths[i] = maxInt(widths[i], displayWidth(col))
		}
	}
	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, formatRowLine(header, widths, rightAlign))
	lines = append(lines, formatSepLine(widths))
	for _, row := range rows {
		lines = append(lines, formatRowLine(row, widths, rightAlign))
	}
	return lines
}

func formatRowLine(cols []string, widths []int, rightAlign []bool) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		if i < len(rightAlign) && rightAlign[i] {
			parts[i] = padLeft(col, widths[i])
		} else {
			parts[i] = padRight(col, widths[i])
		}
	}
	return strings.Join(parts, "  ")
}

func formatSepLine(widths []int) string {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("-", width)
	}
	return strings.Join(parts, "  ")
}

func padRight(value string, width int) string {
	if width <= 0 {
		return value
	}
	if displayWidth(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-displayWidth(value))
}

func padLeft(value string, width int) string {
	if width <= 0 {
		return value
	}
	if displayWidth(value) >= width {
		return value
	}
	return strings.Repeat(" ", width-displayWidth(value)) + value
}
