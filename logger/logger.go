package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes a zap logger with the specified log level and log file.
func InitLogger(level string, logFile string) (*zap.Logger, error) {
	// Define encoder configuration
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// Console encoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)

	// File encoder (JSON format)
	fileEncoder := zapcore.NewJSONEncoder(encoderCfg)

	// Open log file
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file %s: %v\n", logFile, err)
		return nil, err
	}

	// Define log levels
	var zapLevel zapcore.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		zapLevel = zapcore.DebugLevel
	case "INFO":
		zapLevel = zapcore.InfoLevel
	case "WARN", "WARNING":
		zapLevel = zapcore.WarnLevel
	case "ERROR":
		zapLevel = zapcore.ErrorLevel
	default:
		// Default to INFO level
		zapLevel = zapcore.InfoLevel
	}

	// Define write syncers
	consoleWS := zapcore.Lock(os.Stdout)
	fileWS := zapcore.AddSync(file)

	// Create individual cores
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWS, zapLevel)
	fileCore := zapcore.NewCore(fileEncoder, fileWS, zapLevel)

	// Combine cores using Tee
	combinedCore := zapcore.NewTee(consoleCore, fileCore)

	// Build the logger
	logger := zap.New(combinedCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
