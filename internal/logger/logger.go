package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.SugaredLogger

// Init initializes the global logger with file and console output.
func Init(logDir string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Console encoder
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	// File encoder (JSON for structured logs)
	fileEncoderCfg := zap.NewProductionEncoderConfig()
	fileEncoderCfg.TimeKey = "time"
	fileEncoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderCfg)

	// File writer with rotation
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "goserd.log"),
		MaxSize:    50, // MB
		MaxBackups: 5,
		MaxAge:     7, // days
		Compress:   true,
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), zapcore.InfoLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	globalLogger = logger.Sugar()
	return nil
}

// Get returns the global sugared logger.
func Get() *zap.SugaredLogger {
	if globalLogger == nil {
		// Fallback to a no-op logger if not initialized
		l, _ := zap.NewDevelopment()
		globalLogger = l.Sugar()
	}
	return globalLogger
}

// Sync flushes any buffered log entries.
func Sync() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}
