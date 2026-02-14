package config

import "time"

// ServiceConfig defines a managed service's configuration.
type ServiceConfig struct {
	Name         string            `yaml:"name"          json:"name"`
	Command      string            `yaml:"command"       json:"command"`
	Args         []string          `yaml:"args"          json:"args,omitempty"`
	WorkingDir   string            `yaml:"working_dir"   json:"working_dir,omitempty"`
	Env          map[string]string `yaml:"env"           json:"env,omitempty"`
	AutoStart    bool              `yaml:"auto_start"    json:"auto_start"`
	AutoRestart  bool              `yaml:"auto_restart"  json:"auto_restart"`
	MaxRestarts  int               `yaml:"max_restarts"  json:"max_restarts"`
	RestartDelay time.Duration     `yaml:"restart_delay" json:"restart_delay"`
	StopSignal   string            `yaml:"stop_signal"   json:"stop_signal"`
	StopTimeout  time.Duration     `yaml:"stop_timeout"  json:"stop_timeout"`
	LogFile      string            `yaml:"log_file"      json:"log_file"`
	DependsOn    []string          `yaml:"depends_on"    json:"depends_on,omitempty"`
	HealthCheck  *HealthCheckConfig `yaml:"health_check" json:"health_check,omitempty"`
}

// Validate checks the service configuration for required fields and applies defaults.
func (c *ServiceConfig) Validate() error {
	if c.Name == "" {
		return ErrMissingName
	}
	if c.Command == "" {
		return ErrMissingCommand
	}
	// Apply defaults
	if c.MaxRestarts == 0 {
		c.MaxRestarts = 5
	}
	if c.RestartDelay == 0 {
		c.RestartDelay = 5 * time.Second
	}
	if c.StopSignal == "" {
		c.StopSignal = "SIGTERM"
	}
	if c.StopTimeout == 0 {
		c.StopTimeout = 10 * time.Second
	}
	if c.LogFile == "" {
		c.LogFile = "auto"
	}
	return nil
}

// Errors for service configuration validation.
var (
	ErrMissingName    = &ConfigError{Field: "name", Message: "service name is required"}
	ErrMissingCommand = &ConfigError{Field: "command", Message: "command is required"}
)

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error: " + e.Field + ": " + e.Message
}
