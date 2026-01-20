package main

import (
	"context"
	"errors"
	"fmt"
	"html"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/zergo0/simple-system-monitor/internal/alerts"
	"github.com/zergo0/simple-system-monitor/internal/config"
	"github.com/zergo0/simple-system-monitor/internal/monitor"
	"github.com/zergo0/simple-system-monitor/internal/telegram"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Warn("dotenv load failed", zap.Error(err))
	}

	cfg := config.Load()
	if cfg.LogInterval < time.Second {
		logger.Warn("log interval too small, defaulting to 1s", zap.Duration("interval", cfg.LogInterval))
		cfg.LogInterval = time.Second
	}
	if cfg.CPUAlertWindow < 0 {
		logger.Warn("cpu alert window invalid, disabling delay", zap.Duration("cpu_alert_window", cfg.CPUAlertWindow))
		cfg.CPUAlertWindow = 0
	}
	if cfg.MemAlertWindow < 0 {
		logger.Warn("mem alert window invalid, disabling delay", zap.Duration("mem_alert_window", cfg.MemAlertWindow))
		cfg.MemAlertWindow = 0
	}
	if cfg.DiskAlertWindow < 0 {
		logger.Warn("disk alert window invalid, disabling delay", zap.Duration("disk_alert_window", cfg.DiskAlertWindow))
		cfg.DiskAlertWindow = 0
	}

	hostname, err := os.Hostname()
	if err != nil {
		logger.Warn("hostname lookup failed", zap.Error(err))
		hostname = "unknown"
	}
	displayName := hostname
	if cfg.SystemName != "" {
		displayName = fmt.Sprintf("%s (%s)", cfg.SystemName, hostname)
	}

	ctx, stop := signal.NotifyContext(context.Background(), signalList()...)
	defer stop()

	var telegramClient *telegram.Client
	if cfg.TelegramToken != "" && cfg.TelegramChatID != "" {
		telegramClient = telegram.New(cfg.TelegramToken, cfg.TelegramChatID)
	} else {
		logger.Warn("telegram disabled: missing token or chat id")
	}

	sendTelegramMetrics := telegramClient != nil && cfg.TelegramInterval > 0
	if telegramClient != nil && cfg.TelegramInterval <= 0 {
		logger.Warn("telegram metrics disabled: interval <= 0", zap.Duration("telegram_interval", cfg.TelegramInterval))
	}

	alertState := alerts.NewState()

	now := time.Now()
	if err := runOnce(ctx, logger, telegramClient, displayName, cfg, alertState, now, sendTelegramMetrics); err != nil {
		logger.Error("initial run failed", zap.Error(err))
	}

	nextTelegramAt := time.Time{}
	if sendTelegramMetrics {
		nextTelegramAt = time.Now().Add(cfg.TelegramInterval)
	}

	ticker := time.NewTicker(cfg.LogInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("shutdown", zap.String("reason", ctx.Err().Error()))
			return
		case <-ticker.C:
			now = time.Now()
			sendNow := false
			if sendTelegramMetrics && now.After(nextTelegramAt) {
				sendNow = true
				nextTelegramAt = now.Add(cfg.TelegramInterval)
			}
			if err := runOnce(ctx, logger, telegramClient, displayName, cfg, alertState, now, sendNow); err != nil {
				logger.Error("run failed", zap.Error(err))
			}
		}
	}
}

func runOnce(ctx context.Context, logger *zap.Logger, telegramClient *telegram.Client, hostname string, cfg config.Config, alertState *alerts.AlertState, now time.Time, sendTelegramMetrics bool) error {
	metrics, err := monitor.Collect(ctx, logger, hostname, monitor.FilterConfig{
		MountInclude:  cfg.MountInclude,
		MountExclude:  cfg.MountExclude,
		FstypeExclude: cfg.FstypeExclude,
	})
	if err != nil {
		return err
	}

	logger.Info("system metrics",
		zap.String("hostname", metrics.Hostname),
		zap.Float64("cpu_percent", metrics.CPUPercent),
		zap.Float64("mem_percent", metrics.MemPercent),
		zap.Any("disks", metrics.Disks),
	)

	if sendTelegramMetrics && telegramClient != nil {
		if err := telegramClient.SendHTMLMessage(ctx, monitor.FormatMetricsHTML(metrics)); err != nil {
			logger.Warn("telegram metrics send failed", zap.Error(err))
		}
	}

	alertsList := alerts.Check(metrics, alerts.Thresholds{
		CPUThreshold:    cfg.CPUThreshold,
		CPUAlertWindow:  cfg.CPUAlertWindow,
		MemThreshold:    cfg.MemThreshold,
		MemAlertWindow:  cfg.MemAlertWindow,
		DiskThreshold:   cfg.DiskThreshold,
		DiskAlertWindow: cfg.DiskAlertWindow,
	}, alertState, now)
	if len(alertsList) > 0 {
		logger.Warn("alerts triggered", zap.String("hostname", metrics.Hostname), zap.Strings("alerts", alertsList))
		if telegramClient != nil {
			alertText := formatAlertHTML(metrics.Hostname, alertsList)
			if err := telegramClient.SendHTMLMessage(ctx, alertText); err != nil {
				logger.Warn("telegram alert send failed", zap.Error(err))
			}
		}
	}

	return nil
}

func formatAlertHTML(hostname string, alertsList []string) string {
	var b strings.Builder
	_, _ = b.WriteString("<b>🚨 ALERT</b> ")
	_, _ = b.WriteString(html.EscapeString(hostname))
	if len(alertsList) == 0 {
		return b.String()
	}
	b.WriteString("\n<pre>\n")
	_, _ = b.WriteString(html.EscapeString(strings.Join(alertsList, "\n")))
	b.WriteString("\n</pre>")
	return b.String()
}

func signalList() []os.Signal {
	signals := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		signals = append(signals, syscall.SIGTERM)
	}
	return signals
}
