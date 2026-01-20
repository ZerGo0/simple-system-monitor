# Run as a Linux service (systemd)

This guide sets up Simple System Monitor as a systemd service.

## 1) Install the binary

Use the latest release (example for linux-amd64):

```bash
sudo curl -L -o /usr/local/bin/simple-system-monitor https://github.com/ZerGo0/simple-system-monitor/releases/latest/download/simple-system-monitor-linux-amd64
sudo chmod 0755 /usr/local/bin/simple-system-monitor
```

## 2) Create a service user and working directory

```bash
sudo useradd --system --home /opt/simple-system-monitor --shell /usr/sbin/nologin simple-system-monitor
sudo mkdir -p /opt/simple-system-monitor
sudo chown simple-system-monitor:simple-system-monitor /opt/simple-system-monitor
```

## 3) Configure environment

Create `/opt/simple-system-monitor/.env`:

```bash
sudo tee /opt/simple-system-monitor/.env >/dev/null <<'ENV'
TELEGRAM_BOT_TOKEN=your-token
TELEGRAM_CHAT_ID=your-chat-id
ENV
sudo chown simple-system-monitor:simple-system-monitor /opt/simple-system-monitor/.env
sudo chmod 0600 /opt/simple-system-monitor/.env
```

The service uses `/opt/simple-system-monitor` as its working directory, so `.env` is loaded automatically.

## 4) Create the systemd unit

Create `/etc/systemd/system/simple-system-monitor.service`:

```ini
[Unit]
Description=Simple System Monitor
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=simple-system-monitor
Group=simple-system-monitor
WorkingDirectory=/opt/simple-system-monitor
ExecStart=/usr/local/bin/simple-system-monitor
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

## 5) Enable and start

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now simple-system-monitor
```

## 6) Check status and logs

```bash
systemctl status simple-system-monitor
journalctl -u simple-system-monitor -f
```

## Updating the binary

Replace `/usr/local/bin/simple-system-monitor` with the new release and restart:

```bash
sudo systemctl restart simple-system-monitor
```