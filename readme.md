# System Monitor
 
A self-hosted Linux system monitoring dashboard. Monitor CPU, memory, and processes in real-time from your browser.
 
## Features
 
- 📊 Live CPU, memory, and disk metrics via WebSocket
- 🔔 Configurable alert thresholds with email notifications
- 🧠 Smart Heal — auto-kills anomalous processes using rolling average detection
- ⚙️ System Optimizer — one-click cleanup of CPU/memory hogs and zombies
- 🖥️ Process Manager — monitor, pin, stop, and restart processes
- 📈 Historical Insights — resource trends and alert history
- 🌙 Dark mode
## Quick Start
 
**1. Download the binary**
 
Grab `system-monitor` from the [Releases](../../releases) page.
 
**2. Make it executable**
 
```bash
chmod +x system-monitor
```
 
**3. Run it**
 
```bash
./system-monitor
```
 
**4. Open your browser**
 
Navigate to `http://localhost:8080` and register your account.
 
## Optional Configuration
 
Create a `.env` file next to the binary to enable additional features:
 
```
# Persist sessions across server restarts
JWT_SECRET=your_long_random_secret
 
# Email alert notifications
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@gmail.com
SMTP_PASSWORD=your_app_password
SMTP_FROM=your@gmail.com
```
 
Without a `.env`, the app works fine — sessions reset on restart and email alerts are disabled.
 
## Requirements
 
- Linux x86_64
- No dependencies — single binary, batteries included
## Data
 
Your data is stored locally in `monitor.db` in the same directory as the binary. Nothing is sent externally.
 
## Roadmap
 
- [ ] System tray integration
- [ ] Desktop notifications
- [ ] API pagination
- [ ] Windows/macOS support

## v1.1 Notes (EC2 Deployment):
- Added system action log in the front end.
- Removed hardcoding for process start and restart. This means that ANY process can be started, healed and restarted that doesn't have dependency processes. For example, GUI apps that have multiple other processes tied to it CANNOT be restarted as there would be multiple cmds needed to trigger a restart for the child processes. However, for the end user experience, it would not have an impact since processes likely to be put in the managed list is server related and service processes like nginx. 
- Updated CORS config to allow requests from the d3 hosted frontend.
- Updated the frontend axios baseURL to point to the EC2 public IP.