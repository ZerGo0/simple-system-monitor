# Simple System Monitor

Simple System Monitor: lightweight system monitoring tool with Telegram alerts.

## Requirements
- Go
- Telegram bot token + chat ID

## Configuration
Environment variables or flags:
- `.env` file is loaded automatically when present.
- `TELEGRAM_BOT_TOKEN` / `-telegram-token`
- `TELEGRAM_CHAT_ID` / `-telegram-chat-id`
- `INTERVAL` / `-interval` (log interval, default `1m`)
- `TELEGRAM_INTERVAL` / `-telegram-interval` (metrics send interval, default `168h`)
- `MOUNT_INCLUDE` / `-mount-include` (comma list; only these mounts monitored)
- `MOUNT_EXCLUDE` / `-mount-exclude` (comma list, supports `*` suffix; default `/dev*,/proc*,/sys*,/run*`)
- `FSTYPE_EXCLUDE` / `-fstype-exclude` (comma list; default excludes tmpfs/devtmpfs/etc)
- `CPU_THRESHOLD` / `-cpu-threshold` (percent, default `90`)
- `CPU_ALERT_WINDOW` / `-cpu-alert-window` (duration over threshold before alert, default `5m`)
- `MEM_THRESHOLD` / `-mem-threshold` (percent, default `90`)
- `MEM_ALERT_WINDOW` / `-mem-alert-window` (duration over threshold before alert, default `5m`)
- `DISK_THRESHOLD` / `-disk-threshold` (percent, default `90`)
- `DISK_ALERT_WINDOW` / `-disk-alert-window` (duration over threshold before alert, default `5m`)

## Run
```bash
export TELEGRAM_BOT_TOKEN="<token>"
export TELEGRAM_CHAT_ID="<chat-id>"
export TELEGRAM_INTERVAL="168h"
export MOUNT_INCLUDE="/,/boot/efi,/media/usb3"

go run ./cmd/simple-system-monitor -interval 1m -telegram-interval 168h -mount-include "/,/boot/efi,/media/usb3" -cpu-threshold 85 -mem-threshold 90 -disk-threshold 92
```

## Build (cross-platform)
```bash
GOOS=linux GOARCH=amd64 go build -o simple-system-monitor ./cmd/simple-system-monitor
GOOS=darwin GOARCH=arm64 go build -o simple-system-monitor ./cmd/simple-system-monitor
GOOS=windows GOARCH=amd64 go build -o simple-system-monitor.exe ./cmd/simple-system-monitor
```
