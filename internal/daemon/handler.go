package daemon

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/BAIGUANGMEI/goser/internal/config"
	"github.com/BAIGUANGMEI/goser/internal/logger"
	"github.com/BAIGUANGMEI/goser/internal/model"
)

func (s *Server) setupRoutes() {
	// Logging middleware
	s.router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log := logger.Get()
		log.Debugf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	})

	// CORS middleware for Wails dev mode
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := s.router.Group("/api")
	{
		// Daemon status
		api.GET("/daemon/status", s.handleDaemonStatus)

		// Services
		api.GET("/services", s.handleListServices)
		api.GET("/services/:name", s.handleGetService)
		api.POST("/services", s.handleCreateService)
		api.PUT("/services/:name", s.handleUpdateService)
		api.DELETE("/services/:name", s.handleDeleteService)

		// Service actions
		api.POST("/services/:name/start", s.handleStartService)
		api.POST("/services/:name/stop", s.handleStopService)
		api.POST("/services/:name/restart", s.handleRestartService)

		// Logs
		api.GET("/services/:name/logs", s.handleGetLogs)
	}

	// WebSocket
	s.router.GET("/ws", s.handleWebSocket)
}

// --- Daemon ---

func (s *Server) handleDaemonStatus(c *gin.Context) {
	total, running, stopped, failed := s.mgr.Stats()
	uptime := time.Since(s.startedAt)

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: model.DaemonStatus{
			Running:      true,
			PID:          0, // Will be set by caller
			StartedAt:    s.startedAt,
			Uptime:       formatDuration(uptime),
			ServiceCount: total,
			RunningCount: running,
			StoppedCount: stopped,
			FailedCount:  failed,
		},
	})
}

// --- Services CRUD ---

func (s *Server) handleListServices(c *gin.Context) {
	services := s.mgr.ListServices()
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    services,
	})
}

func (s *Server) handleGetService(c *gin.Context) {
	name := c.Param("name")
	info, err := s.mgr.GetServiceInfo(name)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    info,
	})
}

func (s *Server) handleCreateService(c *gin.Context) {
	var svc config.ServiceConfig
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	if err := s.mgr.AddService(&svc); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "service created",
	})
}

func (s *Server) handleUpdateService(c *gin.Context) {
	name := c.Param("name")
	var svc config.ServiceConfig
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}
	svc.Name = name

	if err := s.mgr.UpdateService(&svc); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "service updated",
	})
}

func (s *Server) handleDeleteService(c *gin.Context) {
	name := c.Param("name")
	if err := s.mgr.RemoveService(name); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "service removed",
	})
}

// --- Service Actions ---

func (s *Server) handleStartService(c *gin.Context) {
	name := c.Param("name")
	if err := s.mgr.StartService(name); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "service started",
	})
}

func (s *Server) handleStopService(c *gin.Context) {
	name := c.Param("name")
	if err := s.mgr.StopService(name); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "service stopped",
	})
}

func (s *Server) handleRestartService(c *gin.Context) {
	name := c.Param("name")
	if err := s.mgr.RestartService(name); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "service restarted",
	})
}

// --- Logs ---

func (s *Server) handleGetLogs(c *gin.Context) {
	name := c.Param("name")
	nStr := c.DefaultQuery("n", "100")
	n, _ := strconv.Atoi(nStr)
	if n <= 0 {
		n = 100
	}

	logs, err := s.mgr.GetServiceLogs(name, n)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    logs,
	})
}

// --- WebSocket ---

func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Get().Errorf("websocket upgrade: %v", err)
		return
	}

	s.wsMu.Lock()
	s.wsClients[conn] = true
	s.wsMu.Unlock()

	// Keep connection alive, remove on disconnect
	defer func() {
		s.wsMu.Lock()
		delete(s.wsClients, conn)
		s.wsMu.Unlock()
		_ = conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return strconv.Itoa(days) + "d " + strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m"
	}
	if hours > 0 {
		return strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m " + strconv.Itoa(seconds) + "s"
	}
	if minutes > 0 {
		return strconv.Itoa(minutes) + "m " + strconv.Itoa(seconds) + "s"
	}
	return strconv.Itoa(seconds) + "s"
}
