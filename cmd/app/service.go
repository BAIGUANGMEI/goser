package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/BAIGUANGMEI/goser/internal/client"
	"github.com/BAIGUANGMEI/goser/internal/config"
	"github.com/BAIGUANGMEI/goser/internal/model"
)

// ServiceBridge provides methods exposed to the Wails frontend.
type ServiceBridge struct {
	ctx    context.Context
	client *client.Client
}

// NewServiceBridge creates a new bridge.
func NewServiceBridge() *ServiceBridge {
	return &ServiceBridge{
		client: client.NewDefault(),
	}
}

func (b *ServiceBridge) startup(ctx context.Context) {
	b.ctx = ctx
}

// GetDaemonStatus returns the daemon status.
func (b *ServiceBridge) GetDaemonStatus() (*model.DaemonStatus, error) {
	return b.client.DaemonStatus()
}

// ListServices returns all services.
func (b *ServiceBridge) ListServices() ([]model.ServiceInfo, error) {
	return b.client.ListServices()
}

// GetService returns a single service by name.
func (b *ServiceBridge) GetService(name string) (*model.ServiceInfo, error) {
	return b.client.GetService(name)
}

// StartService starts a service.
func (b *ServiceBridge) StartService(name string) error {
	return b.client.StartService(name)
}

// StopService stops a service.
func (b *ServiceBridge) StopService(name string) error {
	return b.client.StopService(name)
}

// RestartService restarts a service.
func (b *ServiceBridge) RestartService(name string) error {
	return b.client.RestartService(name)
}

// CreateService adds a new service.
func (b *ServiceBridge) CreateService(svc config.ServiceConfig) error {
	return b.client.CreateService(&svc)
}

// UpdateService updates a service.
func (b *ServiceBridge) UpdateService(name string, svc config.ServiceConfig) error {
	return b.client.UpdateService(name, &svc)
}

// DeleteService removes a service.
func (b *ServiceBridge) DeleteService(name string) error {
	return b.client.DeleteService(name)
}

// GetLogs returns recent logs for a service.
func (b *ServiceBridge) GetLogs(name string, n int) ([]model.LogEntry, error) {
	return b.client.GetLogs(name, n)
}

// GetDaemonAddress returns the daemon connection address.
func (b *ServiceBridge) GetDaemonAddress() string {
	cfg := config.DefaultGlobalConfig()
	return cfg.Daemon.Listen
}

// StartDaemon launches the goserd process in the background.
func (b *ServiceBridge) StartDaemon() error {
	// Check if already running
	_, err := b.client.DaemonStatus()
	if err == nil {
		return nil // already running
	}

	// Ensure config dirs exist
	_ = config.EnsureDirs()

	// Find goserd.exe next to this executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}
	daemonPath := filepath.Join(filepath.Dir(exePath), "goserd.exe")
	if _, err := os.Stat(daemonPath); os.IsNotExist(err) {
		return fmt.Errorf("goserd.exe not found at %s", daemonPath)
	}

	cmd := exec.Command(daemonPath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		HideWindow:    true,
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}
	// Detach
	_ = cmd.Process.Release()

	// Wait briefly for it to come up
	for i := 0; i < 10; i++ {
		time.Sleep(300 * time.Millisecond)
		if _, err := b.client.DaemonStatus(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("daemon started but not responding")
}

// StopDaemon stops the goserd process.
func (b *ServiceBridge) StopDaemon() error {
	cfg := config.DefaultGlobalConfig()
	data, err := os.ReadFile(cfg.Daemon.PIDFile)
	if err != nil {
		return fmt.Errorf("daemon not running (no PID file)")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("invalid PID file")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found")
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("kill daemon: %w", err)
	}

	_ = os.Remove(cfg.Daemon.PIDFile)

	// Wait for it to actually stop
	for i := 0; i < 10; i++ {
		time.Sleep(300 * time.Millisecond)
		if _, err := b.client.DaemonStatus(); err != nil {
			return nil // confirmed stopped
		}
	}
	return nil
}
