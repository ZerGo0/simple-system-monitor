# Run as a Linux service (systemd)

This guide sets up Simple System Monitor as a systemd service.

## 1) Install the binary

Use the latest release (example for linux-amd64):

```bash
sudo curl -L -o /usr/local/bin/simple-system-monitor https://github.com/ZerGo0/simple-system-monitor/releases/latest/download/simple-system-monitor-linux-amd64
sudo chmod 0755 /usr/local/bin/simple-system-monitor
```

## 2) Configure environment

Create `/etc/simple-system-monitor.env`, readable by root only:

```bash
sudo install -m 0600 -o root -g root /dev/null /etc/simple-system-monitor.env
sudo tee /etc/simple-system-monitor.env >/dev/null <<'ENV'
TELEGRAM_BOT_TOKEN=your-token
TELEGRAM_CHAT_ID=your-chat-id
ENV
```

systemd reads this file as root and passes the values to the service.

## 3) Create the systemd unit

Create `/etc/systemd/system/simple-system-monitor.service`:

```ini
[Unit]
Description=Simple System Monitor
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
DynamicUser=yes
EnvironmentFile=/etc/simple-system-monitor.env
ExecStart=/usr/local/bin/simple-system-monitor
Restart=on-failure
RestartSec=10s
CapabilityBoundingSet=
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX AF_NETLINK
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
LockPersonality=yes
SystemCallArchitectures=native

[Install]
WantedBy=multi-user.target
```

`DynamicUser=yes` runs the service as a temporary user with a UID from 61184 to 65519. No account is added to `/etc/passwd`, and the UID cannot collide with container users. A static system user often gets UID 999, which is also the user in the MySQL, MariaDB, and Redis images, so their processes and files appear to belong to the monitor. `DynamicUser=yes` also makes the system read-only for the service, gives it a private `/tmp`, and blocks privilege escalation.

## 4) Enable and start

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now simple-system-monitor
```

## 5) Check status and logs

```bash
systemctl status simple-system-monitor
journalctl -u simple-system-monitor -f
```

## Updating the binary

Replace `/usr/local/bin/simple-system-monitor` with the new release and restart:

```bash
sudo systemctl restart simple-system-monitor
```