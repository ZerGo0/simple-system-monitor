package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	TelegramToken    string
	TelegramChatID   string
	SystemName       string
	LogInterval      time.Duration
	TelegramSchedule string
	CPUThreshold     float64
	CPUAlertWindow   time.Duration
	MemThreshold     float64
	MemAlertWindow   time.Duration
	DiskThreshold    float64
	DiskAlertWindow  time.Duration
	MountInclude     []string
	MountExclude     []string
	FstypeExclude    []string
}

func Load() Config {
	return LoadFrom(flag.CommandLine, os.Getenv, os.Args[1:])
}

func LoadFrom(fs *flag.FlagSet, getenv func(string) string, args []string) Config {
	if fs == nil {
		fs = flag.NewFlagSet("simple-system-monitor", flag.ExitOnError)
	}
	if getenv == nil {
		getenv = os.Getenv
	}

	defaultLogInterval := envDuration(getenv, "INTERVAL", time.Minute)
	defaultTelegramSchedule := envString(getenv, "TELEGRAM_SCHEDULE", "0 12 * * 0")
	defaultCPU := envFloat(getenv, "CPU_THRESHOLD", 90)
	defaultCPUWindow := envDuration(getenv, "CPU_ALERT_WINDOW", 5*time.Minute)
	defaultMem := envFloat(getenv, "MEM_THRESHOLD", 90)
	defaultMemWindow := envDuration(getenv, "MEM_ALERT_WINDOW", 5*time.Minute)
	defaultDisk := envFloat(getenv, "DISK_THRESHOLD", 90)
	defaultDiskWindow := envDuration(getenv, "DISK_ALERT_WINDOW", 5*time.Minute)
	defaultToken := envString(getenv, "TELEGRAM_BOT_TOKEN", "")
	defaultChat := envString(getenv, "TELEGRAM_CHAT_ID", "")
	defaultSystemName := envString(getenv, "SYSTEM_NAME", "")
	defaultMountInclude := envString(getenv, "MOUNT_INCLUDE", "")
	defaultMountExclude := envString(getenv, "MOUNT_EXCLUDE", "/dev*,/proc*,/sys*,/run*")
	defaultFstypeExclude := envString(getenv, "FSTYPE_EXCLUDE", "tmpfs,devtmpfs,overlay,proc,sysfs,devpts,cgroup,cgroup2,pstore,securityfs,debugfs,tracefs,configfs,ramfs,hugetlbfs,mqueue,autofs,binfmt_misc,fusectl,efivarfs")

	logInterval := fs.Duration("interval", defaultLogInterval, "metrics log interval")
	telegramSchedule := fs.String("telegram-schedule", defaultTelegramSchedule, "telegram metrics cron schedule (UTC)")
	cpuThreshold := fs.Float64("cpu-threshold", defaultCPU, "cpu usage percent threshold")
	cpuAlertWindow := fs.Duration("cpu-alert-window", defaultCPUWindow, "cpu threshold window before alert")
	memThreshold := fs.Float64("mem-threshold", defaultMem, "memory usage percent threshold")
	memAlertWindow := fs.Duration("mem-alert-window", defaultMemWindow, "memory threshold window before alert")
	diskThreshold := fs.Float64("disk-threshold", defaultDisk, "disk usage percent threshold")
	diskAlertWindow := fs.Duration("disk-alert-window", defaultDiskWindow, "disk threshold window before alert")
	telegramToken := fs.String("telegram-token", defaultToken, "telegram bot token")
	telegramChatID := fs.String("telegram-chat-id", defaultChat, "telegram chat id")
	systemName := fs.String("system-name", defaultSystemName, "custom system name")
	mountInclude := fs.String("mount-include", defaultMountInclude, "comma-separated mountpoints to include (overrides exclude)")
	mountExclude := fs.String("mount-exclude", defaultMountExclude, "comma-separated mountpoints to exclude (supports * suffix)")
	fstypeExclude := fs.String("fstype-exclude", defaultFstypeExclude, "comma-separated filesystem types to exclude")

	if !fs.Parsed() {
		_ = fs.Parse(args)
	}

	return Config{
		TelegramToken:    *telegramToken,
		TelegramChatID:   *telegramChatID,
		SystemName:       strings.TrimSpace(*systemName),
		LogInterval:      *logInterval,
		TelegramSchedule: strings.TrimSpace(*telegramSchedule),
		CPUThreshold:     clampPercent(*cpuThreshold),
		CPUAlertWindow:   *cpuAlertWindow,
		MemThreshold:     clampPercent(*memThreshold),
		MemAlertWindow:   *memAlertWindow,
		DiskThreshold:    clampPercent(*diskThreshold),
		DiskAlertWindow:  *diskAlertWindow,
		MountInclude:     parseList(*mountInclude),
		MountExclude:     parseList(*mountExclude),
		FstypeExclude:    parseListLower(*fstypeExclude),
	}
}

func envString(getenv func(string) string, key string, def string) string {
	if val := getenv(key); val != "" {
		return val
	}
	return def
}

func envFloat(getenv func(string) string, key string, def float64) float64 {
	val := getenv(key)
	if val == "" {
		return def
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return def
	}
	return parsed
}

func envDuration(getenv func(string) string, key string, def time.Duration) time.Duration {
	val := getenv(key)
	if val == "" {
		return def
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return def
	}
	return parsed
}

func clampPercent(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func parseList(value string) []string {
	if strings.EqualFold(strings.TrimSpace(value), "none") {
		return nil
	}
	parts := strings.Split(value, ",")
	list := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		list = append(list, item)
	}
	return list
}

func parseListLower(value string) []string {
	list := parseList(value)
	for i, item := range list {
		list[i] = strings.ToLower(item)
	}
	return list
}
