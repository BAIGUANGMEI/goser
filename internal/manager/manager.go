package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/BAIGUANGMEI/goser/internal/config"
	"github.com/BAIGUANGMEI/goser/internal/logger"
	"github.com/BAIGUANGMEI/goser/internal/model"
)

// EventHandler is a callback function for service events.
type EventHandler func(event model.Event)

// Manager orchestrates all managed service processes.
type Manager struct {
	mu            sync.RWMutex
	processes     map[string]*Process
	collectors    map[string]*logger.Collector
	loader        *config.Loader
	logDir        string
	eventHandlers []EventHandler
	stopCh        chan struct{}
}

// New creates a new process manager.
func New(loader *config.Loader) *Manager {
	globalCfg := loader.GetGlobal()
	return &Manager{
		processes:  make(map[string]*Process),
		collectors: make(map[string]*logger.Collector),
		loader:     loader,
		logDir:     globalCfg.Daemon.LogDir,
		stopCh:     make(chan struct{}),
	}
}

// OnEvent registers an event handler.
func (m *Manager) OnEvent(handler EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventHandlers = append(m.eventHandlers, handler)
}

func (m *Manager) emitEvent(event model.Event) {
	m.mu.RLock()
	handlers := make([]EventHandler, len(m.eventHandlers))
	copy(handlers, m.eventHandlers)
	m.mu.RUnlock()

	for _, h := range handlers {
		go h(event)
	}
}

// LoadAndStart loads all service configs and starts services with auto_start=true.
func (m *Manager) LoadAndStart() error {
	services := m.loader.GetServices()
	for _, svc := range services {
		m.registerService(svc)
	}

	// Start services with auto_start, respecting dependencies
	order := m.resolveDependencies()
	for _, name := range order {
		proc := m.getProcess(name)
		if proc == nil {
			continue
		}
		if proc.Config().AutoStart {
			if err := m.StartService(name); err != nil {
				logger.Get().Errorf("failed to auto-start %s: %v", name, err)
			}
		}
	}

	return nil
}

func (m *Manager) registerService(svc *config.ServiceConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collector := logger.NewCollector(svc.Name, m.logDir, func(entry model.LogEntry) {
		m.emitEvent(model.Event{
			Type:      model.EventServiceLog,
			Service:   svc.Name,
			Message:   entry.Line,
			Data:      entry,
			Timestamp: entry.Timestamp,
		})
	})

	proc := NewProcess(svc, collector)
	m.processes[svc.Name] = proc
	m.collectors[svc.Name] = collector
}

// StartService starts a service by name.
func (m *Manager) StartService(name string) error {
	proc := m.getProcess(name)
	if proc == nil {
		return fmt.Errorf("service %s not found", name)
	}

	if err := proc.Start(); err != nil {
		return err
	}

	proc.ResetRestartCount()

	m.emitEvent(model.Event{
		Type:      model.EventServiceStarted,
		Service:   name,
		Message:   "service started",
		Timestamp: time.Now(),
	})

	// Start the monitor goroutine for auto-restart
	go m.monitor(proc)

	return nil
}

// StopService stops a running service.
func (m *Manager) StopService(name string) error {
	proc := m.getProcess(name)
	if proc == nil {
		return fmt.Errorf("service %s not found", name)
	}

	if err := proc.Stop(); err != nil {
		return err
	}

	m.emitEvent(model.Event{
		Type:      model.EventServiceStopped,
		Service:   name,
		Message:   "service stopped",
		Timestamp: time.Now(),
	})

	return nil
}

// RestartService restarts a service.
func (m *Manager) RestartService(name string) error {
	proc := m.getProcess(name)
	if proc == nil {
		return fmt.Errorf("service %s not found", name)
	}

	// Stop if running
	if proc.State() == model.StateRunning {
		if err := proc.Stop(); err != nil {
			return fmt.Errorf("stop before restart: %w", err)
		}
		// Brief pause
		time.Sleep(500 * time.Millisecond)
	}

	return m.StartService(name)
}

// AddService adds a new service from config.
func (m *Manager) AddService(svc *config.ServiceConfig) error {
	if err := svc.Validate(); err != nil {
		return err
	}

	// Save to disk
	if err := m.loader.SaveService(svc); err != nil {
		return err
	}

	// Register in manager
	m.registerService(svc)

	m.emitEvent(model.Event{
		Type:      model.EventServiceAdded,
		Service:   svc.Name,
		Message:   "service added",
		Timestamp: time.Now(),
	})

	return nil
}

// RemoveService removes a service (stops it first if running).
func (m *Manager) RemoveService(name string) error {
	proc := m.getProcess(name)
	if proc != nil && proc.State() == model.StateRunning {
		if err := proc.Stop(); err != nil {
			return fmt.Errorf("stop before remove: %w", err)
		}
	}

	// Remove config from disk
	if err := m.loader.RemoveService(name); err != nil {
		return err
	}

	// Remove from manager
	m.mu.Lock()
	if c, ok := m.collectors[name]; ok {
		_ = c.Close()
		delete(m.collectors, name)
	}
	delete(m.processes, name)
	m.mu.Unlock()

	m.emitEvent(model.Event{
		Type:      model.EventServiceRemoved,
		Service:   name,
		Message:   "service removed",
		Timestamp: time.Now(),
	})

	return nil
}

// UpdateService updates a service's configuration.
func (m *Manager) UpdateService(svc *config.ServiceConfig) error {
	if err := svc.Validate(); err != nil {
		return err
	}

	proc := m.getProcess(svc.Name)
	if proc != nil {
		proc.UpdateConfig(svc)
	}

	if err := m.loader.SaveService(svc); err != nil {
		return err
	}

	m.emitEvent(model.Event{
		Type:      model.EventServiceUpdated,
		Service:   svc.Name,
		Message:   "service configuration updated",
		Timestamp: time.Now(),
	})

	return nil
}

// GetServiceInfo returns info for a single service.
func (m *Manager) GetServiceInfo(name string) (*model.ServiceInfo, error) {
	proc := m.getProcess(name)
	if proc == nil {
		return nil, fmt.Errorf("service %s not found", name)
	}
	info := proc.Info()
	return &info, nil
}

// ListServices returns info for all services.
func (m *Manager) ListServices() []model.ServiceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []model.ServiceInfo
	for _, proc := range m.processes {
		result = append(result, proc.Info())
	}
	return result
}

// GetServiceLogs returns the last n log lines for a service.
func (m *Manager) GetServiceLogs(name string, n int) ([]model.LogEntry, error) {
	m.mu.RLock()
	collector, ok := m.collectors[name]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return collector.GetLines(n), nil
}

// StopAll stops all running services gracefully.
func (m *Manager) StopAll() {
	close(m.stopCh)

	m.mu.RLock()
	procs := make([]*Process, 0, len(m.processes))
	for _, p := range m.processes {
		procs = append(procs, p)
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, p := range procs {
		if p.State() == model.StateRunning {
			wg.Add(1)
			go func(proc *Process) {
				defer wg.Done()
				_ = proc.Stop()
			}(p)
		}
	}
	wg.Wait()

	// Close collectors
	m.mu.Lock()
	for _, c := range m.collectors {
		_ = c.Close()
	}
	m.mu.Unlock()
}

// Stats returns daemon-level statistics.
func (m *Manager) Stats() (total, running, stopped, failed int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.processes {
		total++
		switch p.State() {
		case model.StateRunning:
			running++
		case model.StateStopped:
			stopped++
		case model.StateFailed:
			failed++
		default:
			stopped++
		}
	}
	return
}

func (m *Manager) getProcess(name string) *Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processes[name]
}

// resolveDependencies does a topological sort of services by depends_on.
func (m *Manager) resolveDependencies() []string {
	services := m.loader.GetServices()

	// Build adjacency
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	for name := range services {
		if _, ok := graph[name]; !ok {
			graph[name] = nil
		}
		if _, ok := inDegree[name]; !ok {
			inDegree[name] = 0
		}
	}
	for name, svc := range services {
		for _, dep := range svc.DependsOn {
			graph[dep] = append(graph[dep], name)
			inDegree[name]++
		}
	}

	// Kahn's algorithm
	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	var order []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		for _, neighbor := range graph[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return order
}
