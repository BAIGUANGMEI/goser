package manager

import (
	"time"

	"github.com/BAIGUANGMEI/goser/internal/logger"
	"github.com/BAIGUANGMEI/goser/internal/model"
)

// monitor watches a process and handles auto-restart logic.
func (m *Manager) monitor(proc *Process) {
	log := logger.Get()
	cfg := proc.Config()

	for {
		// Wait for the process to exit
		<-proc.DoneCh()

		state := proc.State()
		if state == model.StateStopped {
			// Intentionally stopped, don't restart
			log.Infof("monitor: %s stopped intentionally, not restarting", cfg.Name)
			return
		}

		if !cfg.AutoRestart {
			log.Infof("monitor: %s exited, auto_restart is disabled", cfg.Name)
			m.emitEvent(model.Event{
				Type:      model.EventServiceFailed,
				Service:   cfg.Name,
				Message:   "service exited and auto_restart is disabled",
				Timestamp: time.Now(),
			})
			return
		}

		if proc.RestartCount() >= cfg.MaxRestarts {
			log.Warnf("monitor: %s exceeded max_restarts (%d), giving up", cfg.Name, cfg.MaxRestarts)
			m.emitEvent(model.Event{
				Type:      model.EventServiceFailed,
				Service:   cfg.Name,
				Message:   "exceeded max restarts",
				Timestamp: time.Now(),
			})
			return
		}

		proc.IncrementRestartCount()
		delay := cfg.RestartDelay
		log.Infof("monitor: restarting %s in %s (attempt %d/%d)",
			cfg.Name, delay, proc.RestartCount(), cfg.MaxRestarts)

		// Wait before restarting
		select {
		case <-time.After(delay):
		case <-m.stopCh:
			return
		}

		// Restart the process
		if err := proc.Start(); err != nil {
			log.Errorf("monitor: failed to restart %s: %v", cfg.Name, err)
			m.emitEvent(model.Event{
				Type:      model.EventServiceFailed,
				Service:   cfg.Name,
				Message:   "restart failed: " + err.Error(),
				Timestamp: time.Now(),
			})
			return
		}

		m.emitEvent(model.Event{
			Type:      model.EventServiceRestarted,
			Service:   cfg.Name,
			Message:   "service restarted",
			Timestamp: time.Now(),
		})
	}
}
