package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BAIGUANGMEI/goser/internal/config"
	"github.com/BAIGUANGMEI/goser/internal/model"
)

// Client communicates with the GoSer daemon via HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new daemon client.
func New(addr string) *Client {
	return &Client{
		baseURL: "http://" + addr,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewDefault creates a client with the default daemon address.
func NewDefault() *Client {
	cfg := config.DefaultGlobalConfig()
	return New(cfg.Daemon.Listen)
}

// --- Daemon ---

// DaemonStatus returns the daemon status.
func (c *Client) DaemonStatus() (*model.DaemonStatus, error) {
	var resp model.APIResponse
	if err := c.get("/api/daemon/status", &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("daemon error: %s", resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var status model.DaemonStatus
	_ = json.Unmarshal(data, &status)
	return &status, nil
}

// --- Services ---

// ListServices returns all services.
func (c *Client) ListServices() ([]model.ServiceInfo, error) {
	var resp model.APIResponse
	if err := c.get("/api/services", &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("error: %s", resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var services []model.ServiceInfo
	_ = json.Unmarshal(data, &services)
	return services, nil
}

// GetService returns a single service by name.
func (c *Client) GetService(name string) (*model.ServiceInfo, error) {
	var resp model.APIResponse
	if err := c.get("/api/services/"+name, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("error: %s", resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var info model.ServiceInfo
	_ = json.Unmarshal(data, &info)
	return &info, nil
}

// CreateService creates a new service.
func (c *Client) CreateService(svc *config.ServiceConfig) error {
	var resp model.APIResponse
	if err := c.post("/api/services", svc, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("error: %s", resp.Error)
	}
	return nil
}

// UpdateService updates a service configuration.
func (c *Client) UpdateService(name string, svc *config.ServiceConfig) error {
	var resp model.APIResponse
	if err := c.put("/api/services/"+name, svc, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("error: %s", resp.Error)
	}
	return nil
}

// DeleteService removes a service.
func (c *Client) DeleteService(name string) error {
	var resp model.APIResponse
	if err := c.delete("/api/services/"+name, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("error: %s", resp.Error)
	}
	return nil
}

// --- Service Actions ---

// StartService starts a service.
func (c *Client) StartService(name string) error {
	var resp model.APIResponse
	if err := c.post("/api/services/"+name+"/start", nil, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("error: %s", resp.Error)
	}
	return nil
}

// StopService stops a service.
func (c *Client) StopService(name string) error {
	var resp model.APIResponse
	if err := c.post("/api/services/"+name+"/stop", nil, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("error: %s", resp.Error)
	}
	return nil
}

// RestartService restarts a service.
func (c *Client) RestartService(name string) error {
	var resp model.APIResponse
	if err := c.post("/api/services/"+name+"/restart", nil, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("error: %s", resp.Error)
	}
	return nil
}

// --- Logs ---

// GetLogs returns recent logs for a service.
func (c *Client) GetLogs(name string, n int) ([]model.LogEntry, error) {
	url := fmt.Sprintf("/api/services/%s/logs?n=%d", name, n)
	var resp model.APIResponse
	if err := c.get(url, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("error: %s", resp.Error)
	}

	data, _ := json.Marshal(resp.Data)
	var logs []model.LogEntry
	_ = json.Unmarshal(data, &logs)
	return logs, nil
}

// --- HTTP helpers ---

func (c *Client) get(path string, result interface{}) error {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("connect to daemon: %w (is the daemon running?)", err)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) post(path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	resp, err := c.httpClient.Post(c.baseURL+path, "application/json", bodyReader)
	if err != nil {
		return fmt.Errorf("connect to daemon: %w (is the daemon running?)", err)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) put(path string, body interface{}, result interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to daemon: %w (is the daemon running?)", err)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) delete(path string, result interface{}) error {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to daemon: %w (is the daemon running?)", err)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}
