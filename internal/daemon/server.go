package daemon

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/BAIGUANGMEI/goser/internal/config"
	"github.com/BAIGUANGMEI/goser/internal/logger"
	"github.com/BAIGUANGMEI/goser/internal/manager"
	"github.com/BAIGUANGMEI/goser/internal/model"
)

// Server is the daemon HTTP server that exposes the REST API and WebSocket.
type Server struct {
	cfg       *config.GlobalConfig
	loader    *config.Loader
	mgr       *manager.Manager
	router    *gin.Engine
	wsClients map[*websocket.Conn]bool
	wsMu      sync.Mutex
	startedAt time.Time
}

// New creates a new daemon server.
func New(loader *config.Loader) *Server {
	cfg := loader.GetGlobal()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	mgr := manager.New(loader)

	s := &Server{
		cfg:       cfg,
		loader:    loader,
		mgr:       mgr,
		router:    router,
		wsClients: make(map[*websocket.Conn]bool),
		startedAt: time.Now(),
	}

	// Register event handler for WebSocket broadcasting
	mgr.OnEvent(func(event model.Event) {
		s.broadcastEvent(event)
	})

	s.setupRoutes()
	return s
}

// Run starts the HTTP server and blocks until shutdown.
func (s *Server) Run() error {
	log := logger.Get()

	// Write PID file
	if err := s.writePIDFile(); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}
	defer s.removePIDFile()

	// Load and auto-start services
	if err := s.mgr.LoadAndStart(); err != nil {
		log.Errorf("load and start services: %v", err)
	}

	s.mgr.OnEvent(func(event model.Event) {
		if event.Type != model.EventServiceLog {
			log.Infof("event: %s %s - %s", event.Type, event.Service, event.Message)
		}
	})

	// Start HTTP server
	srv := &http.Server{
		Addr:    s.cfg.Daemon.Listen,
		Handler: s.router,
	}

	// Graceful shutdown on signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Infof("daemon listening on %s", s.cfg.Daemon.Listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	log.Info("shutting down daemon...")

	// Stop all services
	s.mgr.StopAll()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("server shutdown: %v", err)
	}

	// Close WebSocket clients
	s.wsMu.Lock()
	for conn := range s.wsClients {
		_ = conn.Close()
	}
	s.wsMu.Unlock()

	log.Info("daemon stopped")
	return nil
}

func (s *Server) writePIDFile() error {
	pid := os.Getpid()
	return os.WriteFile(s.cfg.Daemon.PIDFile, []byte(strconv.Itoa(pid)), 0644)
}

func (s *Server) removePIDFile() {
	_ = os.Remove(s.cfg.Daemon.PIDFile)
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local daemon
	},
}

func (s *Server) broadcastEvent(event model.Event) {
	s.wsMu.Lock()
	defer s.wsMu.Unlock()

	for conn := range s.wsClients {
		if err := conn.WriteJSON(event); err != nil {
			_ = conn.Close()
			delete(s.wsClients, conn)
		}
	}
}
