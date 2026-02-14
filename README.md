# GoSer - Go Service Manager

<p align="center">
  <img src="build/windows/logo.png" width="128" alt="GoSer Logo">
</p>

<p align="center">
  A non-blocking service manager for Windows, built with Go.<br>
  Similar to Linux's <code>systemd</code>, GoSer manages background processes with a modern GUI and CLI.
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#cli-reference">CLI</a> •
  <a href="#gui">GUI</a> •
  <a href="#rest-api">API</a> •
  <a href="#license">License</a>
</p>

---

## Features

- **Process Management** — Start, stop, restart, and monitor background services
- **Auto-restart** — Configurable restart on failure with max retry limits
- **YAML Configuration** — Simple per-service YAML config files
- **Modern GUI** — Desktop application built with Wails + Vue 3 + TailwindCSS
- **CLI** — Full-featured command-line interface
- **Real-time Logs** — View output with WebSocket streaming, content-based error highlighting
- **Service Dependencies** — Topological ordering via `depends_on`
- **Windows Service** — Can run as a native Windows service
- **Health Checks** — HTTP, TCP, and command-based health checks
- **Process Tree Management** — Reliable process tree termination using Windows Job Objects

## Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   GoSer CLI     │    │  GoSer GUI App  │
│  (goser.exe)    │    │ (goser-app.exe) │
└────────┬────────┘    └────────┬────────┘
         │   HTTP REST + WS    │
         └──────────┬──────────┘
                    │
         ┌──────────▼──────────┐
         │   GoSer Daemon      │
         │   (goserd.exe)      │
         │                     │
         │  ┌───────────────┐  │
         │  │Process Manager│  │
         │  └───────┬───────┘  │
         │          │          │
         │  ┌───┐ ┌───┐ ┌───┐ │
         │  │Svc│ │Svc│ │Svc│ │
         │  │ A │ │ B │ │ C │ │
         │  └───┘ └───┘ └───┘ │
         └─────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- [Inno Setup 6](https://jrsoftware.org/isinfo.php) (optional, for installer)

### Build

```powershell
# Build all binaries (daemon, CLI, GUI)
.\build\build.ps1

# Build with installer
.\build\build.ps1 -Installer
```

Or build individually:

```powershell
# Frontend
cd cmd/app/frontend && npm install && npx vite build && cd ../../..

# Daemon
go build -ldflags="-s -w" -o dist/goserd.exe ./cmd/goserd

# CLI
go build -ldflags="-s -w" -o dist/goser.exe ./cmd/goser

# GUI (requires Wails build tags)
go build -tags "desktop,production" -ldflags="-s -w -H=windowsgui" -o dist/goser-app.exe ./cmd/app
```

### Start the Daemon

```powershell
goserd.exe
# or via CLI
goser daemon start
```

### Add a Service

Create a YAML config file:

```yaml
name: my-web-app
command: node
args:
  - server.js
working_dir: "C:/projects/my-app"
env:
  NODE_ENV: production
  PORT: "3000"
auto_start: true
auto_restart: true
max_restarts: 5
restart_delay: 5s
```

```powershell
goser add my-service.yaml
goser start my-web-app
```

### Monitor

```powershell
goser list                  # List all services
goser status my-web-app     # Detailed status
goser logs my-web-app       # View logs
```

### Launch GUI

```powershell
goser-app.exe
```

## CLI Reference

```
goser daemon start          Start the daemon in background
goser daemon stop           Stop the daemon
goser daemon status         Check daemon status

goser list                  List all services with status
goser start <name>          Start a service
goser stop <name>           Stop a service
goser restart <name>        Restart a service
goser status <name>         Detailed service status

goser add <yaml-file>       Add a service from YAML file
goser remove <name>         Remove a service
goser enable <name>         Enable auto-start
goser disable <name>        Disable auto-start

goser logs <name>           View recent logs
goser logs -n 100 <name>    View last 100 lines
```

## GUI

The GUI provides a modern interface with:

- **Dashboard** — Overview of service statistics and daemon status
- **Services** — List, search, filter, create, edit, and manage services
- **Logs** — Real-time log streaming with content-based error highlighting
- **Settings** — Configure daemon connection
- **Daemon Control** — Start/stop daemon directly from the sidebar

## Service Configuration

Service files are stored in `~/.goser/services/<name>.yaml`:

```yaml
name: my-service            # Required: unique service name
command: node               # Required: executable to run
args:                       # Optional: command arguments
  - server.js
  - --port=3000
working_dir: "C:/myapp"    # Optional: working directory
env:                        # Optional: environment variables
  NODE_ENV: production
auto_start: true            # Start when daemon starts
auto_restart: true          # Restart on failure
max_restarts: 5             # Max restart attempts
restart_delay: 5s           # Delay between restarts
stop_timeout: 10s           # Force kill timeout
depends_on:                 # Optional: service dependencies
  - database
health_check:               # Optional: health monitoring
  type: http
  endpoint: "http://localhost:3000/health"
  interval: 30s
  timeout: 5s
```

## Global Configuration

Located at `~/.goser/config.yaml`:

```yaml
daemon:
  listen: "127.0.0.1:9876"
  log_dir: "~/.goser/logs"
  pid_file: "~/.goser/goserd.pid"
  max_log_size: "50MB"
  log_retention: 7
```

## Windows Service

Install as a Windows service for auto-start on boot:

```powershell
# Install (run as Administrator)
goserd.exe -install

# Start via Windows services
sc start GoSerDaemon

# Uninstall
goserd.exe -uninstall
```

## REST API

The daemon exposes a REST API on `127.0.0.1:9876`:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/daemon/status` | Daemon health |
| GET | `/api/services` | List all services |
| GET | `/api/services/:name` | Get service detail |
| POST | `/api/services` | Create service |
| PUT | `/api/services/:name` | Update service |
| DELETE | `/api/services/:name` | Remove service |
| POST | `/api/services/:name/start` | Start service |
| POST | `/api/services/:name/stop` | Stop service |
| POST | `/api/services/:name/restart` | Restart service |
| GET | `/api/services/:name/logs` | Get service logs |
| WS | `/ws` | Real-time events |

## Tech Stack

- **Backend**: Go 1.21+
- **GUI**: [Wails v2](https://wails.io/) + Vue 3 + TypeScript + TailwindCSS v4
- **CLI**: [Cobra](https://github.com/spf13/cobra)
- **HTTP Server**: [Gin](https://github.com/gin-gonic/gin)
- **WebSocket**: [Gorilla WebSocket](https://github.com/gorilla/websocket)
- **Config**: YAML ([gopkg.in/yaml.v3](https://gopkg.in/yaml.v3))
- **Logging**: [Zap](https://github.com/uber-go/zap) + [Lumberjack](https://github.com/natefinish/lumberjack) (rotation)
- **Windows Service**: [kardianos/service](https://github.com/kardianos/service)
- **Process Management**: Windows Job Objects for reliable process tree control

## Project Structure

```
goser/
├── cmd/
│   ├── goserd/              # Daemon entry point
│   ├── goser/               # CLI entry point
│   └── app/                 # Wails GUI entry point
│       └── frontend/        # Vue 3 + TypeScript frontend
├── internal/
│   ├── daemon/              # HTTP server & API handlers
│   ├── manager/             # Process manager core
│   │   ├── process.go       # Single process lifecycle
│   │   ├── manager.go       # Service orchestration
│   │   ├── monitor.go       # Auto-restart monitor
│   │   └── procutil_*.go    # Platform-specific process utils
│   ├── client/              # HTTP client library
│   ├── config/              # Configuration models & loader
│   ├── model/               # Shared data types
│   └── logger/              # Logging & log collection
├── build/
│   ├── build.ps1            # Windows build script
│   ├── windows/             # Installer & icons
│   └── icongen/             # ICO generation utility
├── go.mod
├── Makefile
└── wails.json
```

## License

[MIT](LICENSE)
