package monitor

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"
	"unicode/utf8"

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
	_, _ = fmt.Fprintf(&b, "<b>Simple System Monitor</b>\n<i>%s</i>", host)

	metricHeader := []string{"Metric", "Usage", "St"}
	metricRows := [][]string{
		{"CPU", fmt.Sprintf("%.1f%%", metrics.CPUPercent), statusEmoji(metrics.CPUPercent)},
		{"MEM", fmt.Sprintf("%.1f%%", metrics.MemPercent), statusEmoji(metrics.MemPercent)},
	}
	metricNameWidth := displayWidth(metricHeader[0])
	metricUseWidth := displayWidth(metricHeader[1])
	metricStatusWidth := displayWidth(metricHeader[2])
	for _, row := range metricRows {
		metricNameWidth = maxInt(metricNameWidth, displayWidth(row[0]))
		metricUseWidth = maxInt(metricUseWidth, displayWidth(row[1]))
		metricStatusWidth = maxInt(metricStatusWidth, displayWidth(row[2]))
	}
	metricStatusWidth = maxInt(metricStatusWidth, 2)

	maxMount := maxMountWidth(metrics.Disks, 24)
	diskHeader := []string{"Mount", "Usage", "St", "Used/Total"}
	diskRows := make([][]string, 0, len(metrics.Disks))
	mountWidth := displayWidth(diskHeader[0])
	useWidth := displayWidth(diskHeader[1])
	statusWidth := displayWidth(diskHeader[2])
	sizeWidth := displayWidth(diskHeader[3])

	for _, d := range metrics.Disks {
		totalGiB := bytesToGiB(d.TotalBytes)
		usedGiB := bytesToGiB(d.UsedBytes)
		mount := formatMount(d.Mountpoint, maxMount)
		use := fmt.Sprintf("%.1f%%", d.UsedPercent)
		status := statusEmoji(d.UsedPercent)
		size := fmt.Sprintf("%.1f/%.1fGiB", usedGiB, totalGiB)

		diskRows = append(diskRows, []string{mount, use, status, size})
		mountWidth = maxInt(mountWidth, displayWidth(mount))
		useWidth = maxInt(useWidth, displayWidth(use))
		statusWidth = maxInt(statusWidth, displayWidth(status))
		sizeWidth = maxInt(sizeWidth, displayWidth(size))
	}
	statusWidth = maxInt(statusWidth, 2)

	b.WriteString("\n<pre>\n")
	b.WriteString(tableTop3(metricNameWidth, metricUseWidth, metricStatusWidth))
	b.WriteString(tableRow3(metricNameWidth, metricUseWidth, metricStatusWidth, metricHeader))
	b.WriteString(tableMid3(metricNameWidth, metricUseWidth, metricStatusWidth))
	for i, row := range metricRows {
		b.WriteString(tableRow3(metricNameWidth, metricUseWidth, metricStatusWidth, row))
		if i < len(metricRows)-1 {
			b.WriteString(tableMid3(metricNameWidth, metricUseWidth, metricStatusWidth))
		}
	}
	b.WriteString(tableBottom3(metricNameWidth, metricUseWidth, metricStatusWidth))
	b.WriteString("\n\nDisk\n")
	if len(metrics.Disks) == 0 {
		b.WriteString("none\n</pre>")
		return b.String()
	}
	b.WriteString(tableTop(mountWidth, useWidth, statusWidth, sizeWidth))
	b.WriteString(tableRow(mountWidth, useWidth, statusWidth, sizeWidth, diskHeader))
	b.WriteString(tableMid(mountWidth, useWidth, statusWidth, sizeWidth))
	for i, row := range diskRows {
		row[0] = html.EscapeString(row[0])
		row[3] = html.EscapeString(row[3])
		b.WriteString(tableRow(mountWidth, useWidth, statusWidth, sizeWidth, row))
		if i < len(diskRows)-1 {
			b.WriteString(tableMid(mountWidth, useWidth, statusWidth, sizeWidth))
		}
	}
	b.WriteString(tableBottom(mountWidth, useWidth, statusWidth, sizeWidth))
	b.WriteString("</pre>")
	return b.String()
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

	metricHeader := []string{"Metric", "Usage", "Status"}
	metricRows := [][]string{
		{"CPU", fmt.Sprintf("%.1f%%", metrics.CPUPercent), statusLabel(metrics.CPUPercent)},
		{"MEM", fmt.Sprintf("%.1f%%", metrics.MemPercent), statusLabel(metrics.MemPercent)},
	}
	lines = append(lines, formatTableLines(metricHeader, metricRows, []bool{false, true, false})...)
	lines = append(lines, "", "Disk")
	if len(metrics.Disks) == 0 {
		lines = append(lines, "none")
		return strings.Join(lines, "\n")
	}

	maxMount := maxMountWidth(metrics.Disks, 24)
	diskHeader := []string{"Mount", "Usage", "Status", "Used/Total"}
	diskRows := make([][]string, 0, len(metrics.Disks))
	for _, d := range metrics.Disks {
		totalGiB := bytesToGiB(d.TotalBytes)
		usedGiB := bytesToGiB(d.UsedBytes)
		mount := formatMountPlain(CleanText(d.Mountpoint), maxMount)
		use := fmt.Sprintf("%.1f%%", d.UsedPercent)
		status := statusLabel(d.UsedPercent)
		size := fmt.Sprintf("%.1f/%.1fGiB", usedGiB, totalGiB)
		diskRows = append(diskRows, []string{mount, use, status, size})
	}
	lines = append(lines, formatTableLines(diskHeader, diskRows, []bool{false, true, false, true})...)
	return strings.Join(lines, "\n")
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

func formatMount(mount string, width int) string {
	return formatMountEllipsis(mount, width, "…")
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

func tableTop3(nameW, useW, statusW int) string {
	return fmt.Sprintf("┌%s┬%s┬%s┐\n",
		strings.Repeat("─", nameW+2),
		strings.Repeat("─", useW+2),
		strings.Repeat("─", statusW+2),
	)
}

func tableMid3(nameW, useW, statusW int) string {
	return fmt.Sprintf("├%s┼%s┼%s┤\n",
		strings.Repeat("─", nameW+2),
		strings.Repeat("─", useW+2),
		strings.Repeat("─", statusW+2),
	)
}

func tableBottom3(nameW, useW, statusW int) string {
	return fmt.Sprintf("└%s┴%s┴%s┘",
		strings.Repeat("─", nameW+2),
		strings.Repeat("─", useW+2),
		strings.Repeat("─", statusW+2),
	)
}

func tableRow3(nameW, useW, statusW int, cols []string) string {
	name := padRight(cols[0], nameW)
	use := padLeft(cols[1], useW)
	status := padRight(cols[2], statusW)
	return fmt.Sprintf("│ %s │ %s │ %s │\n", name, use, status)
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
