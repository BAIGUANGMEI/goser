package manager

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/BAIGUANGMEI/goser/internal/config"
	"github.com/BAIGUANGMEI/goser/internal/logger"
	"github.com/BAIGUANGMEI/goser/internal/model"
)

// Process wraps a managed child process with lifecycle control.
type Process struct {
	mu           sync.RWMutex
	config       *config.ServiceConfig
	cmd          *exec.Cmd
	state        model.ServiceState
	pid          int
	exitCode     *int
	startedAt    *time.Time
	stoppedAt    *time.Time
	restartCount int
	lastError    string
	collector    *logger.Collector
	stopCh       chan struct{}
	doneCh       chan struct{}
}

// NewProcess creates a new Process for the given service config.
func NewProcess(cfg *config.ServiceConfig, collector *logger.Collector) *Process {
	return &Process{
		config:    cfg,
		state:     model.StateStopped,
		collector: collector,
	}
}

// Start launches the child process.
func (p *Process) Start() error {
	p.mu.Lock()
	if p.state == model.StateRunning || p.state == model.StateStarting {
		p.mu.Unlock()
		return fmt.Errorf("service %s is already %s", p.config.Name, p.state)
	}
	p.state = model.StateStarting
	p.mu.Unlock()

	log := logger.Get()
	log.Infof("starting service: %s", p.config.Name)

	cmd := exec.Command(p.config.Command, p.config.Args...)

	// Set working directory
	if p.config.WorkingDir != "" {
		cmd.Dir = p.config.WorkingDir
	}

	// Set environment
	if len(p.config.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range p.config.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	// Pipe stdout and stderr to log collector
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		p.setFailed(fmt.Sprintf("stdout pipe: %v", err))
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		p.setFailed(fmt.Sprintf("stderr pipe: %v", err))
		return err
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		p.setFailed(fmt.Sprintf("start: %v", err))
		return fmt.Errorf("start %s: %w", p.config.Name, err)
	}

	now := time.Now()
	p.mu.Lock()
	p.cmd = cmd
	p.pid = cmd.Process.Pid
	p.startedAt = &now
	p.stoppedAt = nil
	p.exitCode = nil
	p.lastError = ""
	p.state = model.StateRunning
	p.stopCh = make(chan struct{})
	p.doneCh = make(chan struct{})
	p.mu.Unlock()

	log.Infof("service %s started with PID %d", p.config.Name, p.pid)

	// Collect logs in background
	go p.collector.Collect(stdout, "stdout")
	go p.collector.Collect(stderr, "stderr")

	// Wait for process to exit in background
	go p.wait()

	return nil
}

// wait monitors the process until it exits.
func (p *Process) wait() {
	defer close(p.doneCh)

	err := p.cmd.Wait()
	now := time.Now()

	p.mu.Lock()
	p.stoppedAt = &now
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			p.exitCode = &code
		}
		// Only set to failed if we didn't intentionally stop it
		if p.state != model.StateStopping {
			p.lastError = err.Error()
			p.state = model.StateFailed
		} else {
			p.state = model.StateStopped
		}
	} else {
		code := 0
		p.exitCode = &code
		if p.state != model.StateStopping {
			p.state = model.StateStopped
		} else {
			p.state = model.StateStopped
		}
	}
	p.pid = 0
	p.mu.Unlock()

	log := logger.Get()
	log.Infof("service %s exited (exit_code=%v, state=%s)", p.config.Name, p.exitCode, p.state)
}

// Stop gracefully stops the child process.
func (p *Process) Stop() error {
	p.mu.Lock()
	if p.state != model.StateRunning {
		p.mu.Unlock()
		return fmt.Errorf("service %s is not running (state=%s)", p.config.Name, p.state)
	}
	p.state = model.StateStopping
	cmd := p.cmd
	p.mu.Unlock()

	log := logger.Get()
	log.Infof("stopping service: %s (PID %d)", p.config.Name, cmd.Process.Pid)

	// Send kill signal (Windows doesn't support SIGTERM well, use Kill)
	if err := cmd.Process.Kill(); err != nil {
		log.Warnf("failed to kill %s: %v", p.config.Name, err)
		return err
	}

	// Wait for the process to finish with a timeout
	select {
	case <-p.doneCh:
		log.Infof("service %s stopped", p.config.Name)
	case <-time.After(p.config.StopTimeout):
		log.Warnf("service %s stop timeout, force killing", p.config.Name)
		_ = cmd.Process.Kill()
	}

	return nil
}

// Info returns the current runtime info for this process.
func (p *Process) Info() model.ServiceInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	info := model.ServiceInfo{
		Name:         p.config.Name,
		State:        p.state,
		PID:          p.pid,
		Command:      p.config.Command,
		Args:         p.config.Args,
		WorkingDir:   p.config.WorkingDir,
		Env:          p.config.Env,
		AutoStart:    p.config.AutoStart,
		AutoRestart:  p.config.AutoRestart,
		RestartCount: p.restartCount,
		StartedAt:    p.startedAt,
		StoppedAt:    p.stoppedAt,
		ExitCode:     p.exitCode,
		Error:        p.lastError,
	}

	if p.state == model.StateRunning && p.startedAt != nil {
		uptime := time.Since(*p.startedAt)
		info.Uptime = formatDuration(uptime)
	}

	return info
}

// State returns the current state of the process.
func (p *Process) State() model.ServiceState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// Config returns the service config for this process.
func (p *Process) Config() *config.ServiceConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// UpdateConfig updates the service configuration.
func (p *Process) UpdateConfig(cfg *config.ServiceConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = cfg
}

// DoneCh returns a channel that is closed when the process exits.
func (p *Process) DoneCh() <-chan struct{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.doneCh == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return p.doneCh
}

// IncrementRestartCount increments the restart counter.
func (p *Process) IncrementRestartCount() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.restartCount++
}

// ResetRestartCount resets the restart counter to zero.
func (p *Process) ResetRestartCount() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.restartCount = 0
}

// RestartCount returns the current restart count.
func (p *Process) RestartCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.restartCount
}

func (p *Process) setFailed(errMsg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = model.StateFailed
	p.lastError = errMsg
	now := time.Now()
	p.stoppedAt = &now
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
