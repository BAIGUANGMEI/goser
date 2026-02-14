package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Loader handles loading and watching configuration files.
type Loader struct {
	mu       sync.RWMutex
	global   *GlobalConfig
	services map[string]*ServiceConfig
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	return &Loader{
		global:   DefaultGlobalConfig(),
		services: make(map[string]*ServiceConfig),
	}
}

// LoadGlobal loads the global configuration from file.
// If the file does not exist, defaults are used.
func (l *Loader) LoadGlobal() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	cfgPath := filepath.Join(GoserHome(), "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Use defaults and write them out
			return l.writeDefaultGlobal(cfgPath)
		}
		return fmt.Errorf("read global config: %w", err)
	}

	cfg := DefaultGlobalConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse global config: %w", err)
	}
	l.global = cfg
	return nil
}

func (l *Loader) writeDefaultGlobal(path string) error {
	data, err := yaml.Marshal(l.global)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadServices loads all service configurations from the services directory.
func (l *Loader) LoadServices() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	svcDir := ServicesDir()
	entries, err := os.ReadDir(svcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read services dir: %w", err)
	}

	services := make(map[string]*ServiceConfig)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		svc, err := l.loadServiceFile(filepath.Join(svcDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("load service %s: %w", entry.Name(), err)
		}
		services[svc.Name] = svc
	}

	l.services = services
	return nil
}

func (l *Loader) loadServiceFile(path string) (*ServiceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var svc ServiceConfig
	if err := yaml.Unmarshal(data, &svc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := svc.Validate(); err != nil {
		return nil, fmt.Errorf("validate %s: %w", path, err)
	}
	return &svc, nil
}

// GetGlobal returns the current global configuration.
func (l *Loader) GetGlobal() *GlobalConfig {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.global
}

// GetServices returns all loaded service configurations.
func (l *Loader) GetServices() map[string]*ServiceConfig {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[string]*ServiceConfig, len(l.services))
	for k, v := range l.services {
		result[k] = v
	}
	return result
}

// GetService returns a single service configuration by name.
func (l *Loader) GetService(name string) (*ServiceConfig, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	svc, ok := l.services[name]
	return svc, ok
}

// SaveService writes a service configuration to disk and updates the in-memory cache.
func (l *Loader) SaveService(svc *ServiceConfig) error {
	if err := svc.Validate(); err != nil {
		return err
	}

	data, err := yaml.Marshal(svc)
	if err != nil {
		return fmt.Errorf("marshal service config: %w", err)
	}

	path := filepath.Join(ServicesDir(), svc.Name+".yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write service config: %w", err)
	}

	l.mu.Lock()
	l.services[svc.Name] = svc
	l.mu.Unlock()

	return nil
}

// RemoveService removes a service configuration from disk and memory.
func (l *Loader) RemoveService(name string) error {
	path := filepath.Join(ServicesDir(), name+".yaml")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove service config: %w", err)
	}

	l.mu.Lock()
	delete(l.services, name)
	l.mu.Unlock()

	return nil
}
