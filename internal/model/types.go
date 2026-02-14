package model

import "time"

// ServiceState represents the lifecycle state of a managed service.
type ServiceState string

const (
	StateStopped  ServiceState = "stopped"
	StateStarting ServiceState = "starting"
	StateRunning  ServiceState = "running"
	StateStopping ServiceState = "stopping"
	StateFailed   ServiceState = "failed"
)

// ServiceInfo contains runtime information about a managed service.
type ServiceInfo struct {
	Name         string            `json:"name"`
	State        ServiceState      `json:"state"`
	PID          int               `json:"pid,omitempty"`
	Command      string            `json:"command"`
	Args         []string          `json:"args,omitempty"`
	WorkingDir   string            `json:"working_dir,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	AutoStart    bool              `json:"auto_start"`
	AutoRestart  bool              `json:"auto_restart"`
	RestartCount int               `json:"restart_count"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	StoppedAt    *time.Time        `json:"stopped_at,omitempty"`
	Uptime       string            `json:"uptime,omitempty"`
	CPU          float64           `json:"cpu,omitempty"`
	Memory       uint64            `json:"memory,omitempty"`
	ExitCode     *int              `json:"exit_code,omitempty"`
	Error        string            `json:"error,omitempty"`
}

// DaemonStatus contains the status of the daemon process.
type DaemonStatus struct {
	Running      bool      `json:"running"`
	PID          int       `json:"pid"`
	StartedAt    time.Time `json:"started_at"`
	Uptime       string    `json:"uptime"`
	ServiceCount int       `json:"service_count"`
	RunningCount int       `json:"running_count"`
	StoppedCount int       `json:"stopped_count"`
	FailedCount  int       `json:"failed_count"`
}

// EventType represents the type of a service event.
type EventType string

const (
	EventServiceStarted   EventType = "service.started"
	EventServiceStopped   EventType = "service.stopped"
	EventServiceFailed    EventType = "service.failed"
	EventServiceRestarted EventType = "service.restarted"
	EventServiceAdded     EventType = "service.added"
	EventServiceRemoved   EventType = "service.removed"
	EventServiceUpdated   EventType = "service.updated"
	EventServiceLog       EventType = "service.log"
	EventDaemonStarted    EventType = "daemon.started"
	EventDaemonStopping   EventType = "daemon.stopping"
)

// Event represents a real-time event from the daemon.
type Event struct {
	Type      EventType   `json:"type"`
	Service   string      `json:"service,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIResponse is a generic API response wrapper.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// LogEntry represents a single log line from a service.
type LogEntry struct {
	Service   string    `json:"service"`
	Line      string    `json:"line"`
	Stream    string    `json:"stream"` // "stdout" or "stderr"
	Timestamp time.Time `json:"timestamp"`
}
