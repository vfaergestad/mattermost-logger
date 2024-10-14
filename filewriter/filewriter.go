package filewriter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"mattermost-message-monitor/config"
	"mattermost-message-monitor/models"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

type FileWriter struct {
	cfg         *config.Config
	log         *zap.Logger
	currentFile *os.File
	encoder     *json.Encoder
	writer      *bufio.Writer
	timer       *time.Timer
	mu          sync.Mutex
}

// NewFileWriter initializes the FileWriter
func NewFileWriter(cfg *config.Config, log *zap.Logger) (*FileWriter, error) {
	fw := &FileWriter{
		cfg: cfg,
		log: log,
	}

	// Initialize the first file
	if err := fw.rotateFile(); err != nil {
		return nil, err
	}

	// Set the timer for the first rotation
	fw.setNextRotation()

	return fw, nil
}

// setNextRotation calculates the next rotation time and sets the timer
// Supports intervals of 1m, 5m, 1h, 1d, 1w, 1m. Others are defaulted to 1d.
func (fw *FileWriter) setNextRotation() {
	now := time.Now()
	var nextRotation time.Time

	// Calculate the next rotation based on the configured interval
	switch time.Duration(fw.cfg.RotationInterval) {
	case 1 * time.Minute:
		nextRotation = now.Truncate(time.Minute).Add(time.Minute)
	case 5 * time.Minute:
		nextRotation = now.Truncate(5 * time.Minute).Add(5 * time.Minute)
	case 1 * time.Hour:
		nextRotation = now.Truncate(time.Hour).Add(time.Hour)
	case 24 * time.Hour:
		nextRotation = now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	case 7 * 24 * time.Hour: // 1 week
		nextRotation = now.Truncate(7 * 24 * time.Hour).Add(7 * 24 * time.Hour)
	case 30 * 24 * time.Hour: // 1 month
		nextRotation = now.Truncate(30 * 24 * time.Hour).Add(30 * 24 * time.Hour)
	default:
		fw.log.Warn("Unsupported rotation interval; defaulting to 1 day")
		nextRotation = now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	}

	// Set the timer for the next rotation
	durationUntilNextRotation := time.Until(nextRotation)
	fw.timer = time.AfterFunc(durationUntilNextRotation, func() {
		fw.mu.Lock()
		defer fw.mu.Unlock()

		if err := fw.rotateFile(); err != nil {
			fw.log.Error("Failed to rotate file", zap.Error(err))
		}

		// Reset the timer for the next rotation
		fw.setNextRotation()
	})
}

// rotateFile closes the current file and opens a new one with a timestamp
func (fw *FileWriter) rotateFile() error {
	if fw.currentFile != nil {
		// Close existing file
		if err := fw.writer.Flush(); err != nil {
			fw.log.Error("Failed to flush writer during rotation", zap.Error(err))
		}
		if err := fw.currentFile.Close(); err != nil {
			fw.log.Error("Failed to close current file during rotation", zap.Error(err))
		}
	}

	// Generate new file name with timestamp
	timestamp := time.Now().Format("2006-01-02T15-04")
	filename := fmt.Sprintf("%s.%s.json", fw.cfg.OutputFilePrefix, timestamp)
	fullPath := filepath.Join(fw.cfg.OutputDir, filename)

	// Open new file
	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fw.log.Error("Failed to open new output file", zap.String("File", fullPath), zap.Error(err))
		return err
	}

	fw.currentFile = file
	fw.writer = bufio.NewWriter(file)
	fw.encoder = json.NewEncoder(fw.writer)

	fw.log.Info("Rotated to new output file", zap.String("File", fullPath))

	return nil
}

// WriteMessage writes a message to the current file
func (fw *FileWriter) WriteMessage(message models.Message) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.encoder == nil {
		return fmt.Errorf("Encoder is not initialized")
	}

	if err := fw.encoder.Encode(message); err != nil {
		fw.log.Error("Failed to encode message", zap.Error(err))
		return err
	}

	if err := fw.writer.Flush(); err != nil {
		fw.log.Error("Failed to flush writer", zap.Error(err))
		return err
	}

	return nil
}

// Close gracefully shuts down the FileWriter
func (fw *FileWriter) Close() error {
	if fw.timer != nil {
		fw.timer.Stop()
	}
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.writer != nil {
		if err := fw.writer.Flush(); err != nil {
			fw.log.Error("Failed to flush writer on close", zap.Error(err))
		}
	}
	if fw.currentFile != nil {
		if err := fw.currentFile.Close(); err != nil {
			fw.log.Error("Failed to close file on close", zap.Error(err))
		}
	}

	return nil
}
