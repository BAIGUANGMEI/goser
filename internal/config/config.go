package config

import (
	"os"
	"path/filepath"
	"time"
)

// GlobalConfig holds the daemon-level configuration.
type GlobalConfig struct {
	Daemon DaemonConfig `yaml:"daemon"`
}

// DaemonConfig holds daemon-specific configuration.
type DaemonConfig struct {
	Listen       string `yaml:"listen"`
	LogDir       string `yaml:"log_dir"`
	PIDFile      string `yaml:"pid_file"`
	MaxLogSize   string `yaml:"max_log_size"`
	LogRetention int    `yaml:"log_retention"` // days
}

// DefaultGlobalConfig returns a GlobalConfig with sensible defaults.
func DefaultGlobalConfig() *GlobalConfig {
	home := goserHome()
	return &GlobalConfig{
		Daemon: DaemonConfig{
			Listen:       "127.0.0.1:9876",
			LogDir:       filepath.Join(home, "logs"),
			PIDFile:      filepath.Join(home, "goserd.pid"),
			MaxLogSize:   "50MB",
			LogRetention: 7,
		},
	}
}

// HealthCheckConfig configures a health check for a service.
type HealthCheckConfig struct {
	Type     string        `yaml:"type"`     // http | tcp | command
	Endpoint string        `yaml:"endpoint"` // URL or address
	Command  string        `yaml:"command"`  // command to execute
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

// GoserHome returns the path to the goser configuration directory.
func GoserHome() string {
	return goserHome()
}

func goserHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".goser")
}

// ServicesDir returns the path to the services configuration directory.
func ServicesDir() string {
	return filepath.Join(goserHome(), "services")
}

// EnsureDirs creates the goser home and services directories if they don't exist.
func EnsureDirs() error {
	dirs := []string{
		goserHome(),
		ServicesDir(),
		filepath.Join(goserHome(), "logs"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}
