package logger

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BAIGUANGMEI/goser/internal/model"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogCallback is called for each log line collected from a service.
type LogCallback func(entry model.LogEntry)

// Collector captures stdout/stderr from a service process and writes to log files.
type Collector struct {
	serviceName string
	logDir      string
	writer      *lumberjack.Logger
	callback    LogCallback
	mu          sync.Mutex
	lines       []model.LogEntry
	maxLines    int
}

// NewCollector creates a new log collector for a service.
func NewCollector(serviceName, logDir string, callback LogCallback) *Collector {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		Get().Errorf("failed to create log dir %s: %v", logDir, err)
	}

	return &Collector{
		serviceName: serviceName,
		logDir:      logDir,
		writer: &lumberjack.Logger{
			Filename:   filepath.Join(logDir, serviceName+".log"),
			MaxSize:    50, // MB
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
		callback: callback,
		maxLines: 1000, // Keep last 1000 lines in memory
	}
}

// Writer returns an io.Writer that can be connected to process stdout/stderr.
func (c *Collector) Writer() io.Writer {
	return c.writer
}

// Collect reads from a reader (stdout or stderr) line by line.
func (c *Collector) Collect(r io.Reader, stream string) {
	scanner := bufio.NewScanner(r)
	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		entry := model.LogEntry{
			Service:   c.serviceName,
			Line:      line,
			Stream:    stream,
			Timestamp: time.Now(),
		}

		// Write to file
		_, _ = c.writer.Write([]byte(entry.Timestamp.Format(time.RFC3339) + " [" + stream + "] " + line + "\n"))

		// Store in memory ring buffer
		c.mu.Lock()
		c.lines = append(c.lines, entry)
		if len(c.lines) > c.maxLines {
			c.lines = c.lines[len(c.lines)-c.maxLines:]
		}
		c.mu.Unlock()

		// Notify callback
		if c.callback != nil {
			c.callback(entry)
		}
	}
}

// GetLines returns the last n log lines from memory.
func (c *Collector) GetLines(n int) []model.LogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n <= 0 || n > len(c.lines) {
		n = len(c.lines)
	}
	start := len(c.lines) - n
	result := make([]model.LogEntry, n)
	copy(result, c.lines[start:])
	return result
}

// Close closes the log writer.
func (c *Collector) Close() error {
	return c.writer.Close()
}
