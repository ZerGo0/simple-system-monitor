package config

import (
	"flag"
	"io"
	"testing"
	"time"
)

func TestLoadFromDefaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	cfg := LoadFrom(fs, func(string) string { return "" }, []string{})

	if cfg.LogInterval != time.Minute {
		t.Fatalf("expected log interval 1m, got %s", cfg.LogInterval)
	}
	if cfg.TelegramSchedule != "0 12 * * 0" {
		t.Fatalf("expected telegram schedule default, got %s", cfg.TelegramSchedule)
	}
	if cfg.CPUThreshold != 90 {
		t.Fatalf("expected cpu threshold 90, got %.1f", cfg.CPUThreshold)
	}
	if cfg.CPUAlertWindow != 5*time.Minute {
		t.Fatalf("expected cpu alert window 5m, got %s", cfg.CPUAlertWindow)
	}
	if cfg.MemAlertWindow != 5*time.Minute {
		t.Fatalf("expected mem alert window 5m, got %s", cfg.MemAlertWindow)
	}
	if cfg.DiskAlertWindow != 5*time.Minute {
		t.Fatalf("expected disk alert window 5m, got %s", cfg.DiskAlertWindow)
	}
	if len(cfg.MountExclude) == 0 {
		t.Fatalf("expected default mount exclude list")
	}
	if len(cfg.FstypeExclude) == 0 {
		t.Fatalf("expected default fstype exclude list")
	}
}

func TestLoadFromEnvAndArgs(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	env := map[string]string{
		"CPU_THRESHOLD":      "200",
		"CPU_ALERT_WINDOW":   "2m",
		"MEM_THRESHOLD":      "50",
		"DISK_THRESHOLD":     "70",
		"MOUNT_INCLUDE":      "none",
		"FSTYPE_EXCLUDE":     "TmpFS,PROC",
		"TELEGRAM_BOT_TOKEN": "token",
		"TELEGRAM_CHAT_ID":   "chat",
		"TELEGRAM_SCHEDULE":  "0 12 * * 1",
	}
	cfg := LoadFrom(fs, func(key string) string { return env[key] }, []string{"-interval", "30s"})

	if cfg.LogInterval != 30*time.Second {
		t.Fatalf("expected log interval 30s, got %s", cfg.LogInterval)
	}
	if cfg.TelegramSchedule != "0 12 * * 1" {
		t.Fatalf("expected telegram schedule from env, got %s", cfg.TelegramSchedule)
	}
	if cfg.CPUThreshold != 100 {
		t.Fatalf("expected cpu threshold clamped to 100, got %.1f", cfg.CPUThreshold)
	}
	if cfg.CPUAlertWindow != 2*time.Minute {
		t.Fatalf("expected cpu alert window 2m, got %s", cfg.CPUAlertWindow)
	}
	if cfg.MemThreshold != 50 {
		t.Fatalf("expected mem threshold 50, got %.1f", cfg.MemThreshold)
	}
	if cfg.DiskThreshold != 70 {
		t.Fatalf("expected disk threshold 70, got %.1f", cfg.DiskThreshold)
	}
	if cfg.MountInclude != nil {
		t.Fatalf("expected mount include nil for 'none'")
	}
	if len(cfg.FstypeExclude) != 2 || cfg.FstypeExclude[0] != "tmpfs" || cfg.FstypeExclude[1] != "proc" {
		t.Fatalf("expected fstype exclude lowercased, got %#v", cfg.FstypeExclude)
	}
	if cfg.TelegramToken != "token" || cfg.TelegramChatID != "chat" {
		t.Fatalf("expected telegram credentials from env")
	}
}
