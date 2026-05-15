# System Monitor

A self-hosted system resource manager for Linux. View live metrics, control processes, set alerts, and let smart heal keep your system healthy — all from a clean web dashboard.

Built with **Go**, **PostgreSQL**, and **React**. Deployable in one command via Docker.

---

## Features

### Live Dashboard
Real-time CPU, memory, disk, and network metrics pushed to your browser every 5 seconds over WebSocket. No page refreshes, no polling.

### Process Control
A dedicated process manager to view, start, stop, kill, or restart running processes directly from the dashboard — no SSH required.

### Smart Heal
A toggleable auto-healing engine that monitors system health and automatically intervenes when a process is consuming too much CPU or degrading overall system performance.

### Historical Insights
Browse everything your system went through over time. Metrics are retained for 5 days before automatic cleanup. Historical insights support summarized views over any range within that window.

### Alerts & Notifications
Set custom CPU and memory thresholds. Get notified by email the moment your system breaches them. Fully configurable through the settings panel.

---

## Who Is This For

- **Indie and freelance developers** running a VPS or personal server
- **College and grad students** deploying their first real project
- Anyone who wants visibility into their Linux server without the complexity of Grafana/Prometheus or the cost of Datadog

If you've ever found out your server was struggling because a user complained — this is for you.

---

## Why Not Grafana or Prometheus?

| | System Monitor | Grafana + Prometheus | Datadog |
|---|---|---|---|
| Setup time | ~2 minutes | Half a day | Hours + billing |
| Self-hosted | ✅ | ✅ | ❌ |
| Free | ✅ | ✅ | ❌ |
| Process control | ✅ | ❌ | ❌ |
| Auto-heal | ✅ | ❌ | ❌ |
| Docker deploy | ✅ | Complex | N/A |

---

## Getting Started

### With Docker (recommended)

```bash
git clone https://github.com/you/system-monitor
cd system-monitor
docker compose up
```

Open `http://localhost:8080` in your browser. That's it — no Go, no PostgreSQL setup, no manual migrations.

### Without Docker

**Prerequisites:** Go 1.22+, PostgreSQL 15+

```bash
git clone https://github.com/you/system-monitor
cd system-monitor

# Set up environment
cp .env.example .env
# Edit .env with your DB credentials and JWT secret

# Run (migrations run automatically on startup)
go run ./cmd/server
```

---

## Configuration

All configuration lives in `.env`:

```env
JWT_SECRET=your-secret-key-here
DB_HOST=localhost
DB_USER=monitor_user
DB_PASSWORD=yourpassword
DB_NAME=system_monitor
SMTP_HOST=smtp.example.com
SMTP_USER=you@example.com
SMTP_PASSWORD=yourpassword
```

---

## API

All endpoints under `/api/v1` require a JWT token via `Authorization: Bearer <token>`.

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /auth/register | Create account |
| POST | /auth/login | Get JWT token |
| POST | /api/v1/auth/logout | Invalidate session |
| GET | /api/v1/stats | Current metrics snapshot |
| GET | /api/v1/stats/history | Historical metrics (`?limit`, `?from`, `?to`) |
| GET | /api/v1/ws | WebSocket live feed |
| GET | /api/v1/alerts | Active alerts |
| GET | /api/v1/processes | Running processes |

---

## Tech Stack

- **Backend:** Go, Gin, sqlx, gopsutil, golang-migrate
- **Database:** PostgreSQL
- **Frontend:** React
- **Auth:** Stateful JWT (sessions stored in DB)
- **Real-time:** WebSocket (gorilla/websocket)
- **Deploy:** Docker + Docker Compose

---

## License

MIT — free to use, modify, and distribute.