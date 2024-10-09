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
	cfg            *config.Config
	log            *zap.Logger
	mu             sync.Mutex
	currentFile    *os.File
	encoder        *json.Encoder
	writer         *bufio.Writer
	rotationTicker *time.Ticker
	done           chan bool
}

// NewFileWriter initializes the FileWriter
func NewFileWriter(cfg *config.Config, log *zap.Logger) (*FileWriter, error) {
	fw := &FileWriter{
		cfg:  cfg,
		log:  log,
		done: make(chan bool),
	}

	// Initialize the first file
	if err := fw.rotateFile(); err != nil {
		return nil, err
	}

	// Start the rotation ticker
	rotationDuration := time.Duration(cfg.RotationInterval)
	fw.rotationTicker = time.NewTicker(rotationDuration)
	go fw.handleRotation()

	return fw, nil
}

// handleRotation listens for ticker events and rotates the file
func (fw *FileWriter) handleRotation() {
	for {
		select {
		case <-fw.rotationTicker.C:
			fw.mu.Lock()
			if err := fw.rotateFile(); err != nil {
				fw.log.Error("Failed to rotate file", zap.Error(err))
			}
			fw.mu.Unlock()
		case <-fw.done:
			fw.rotationTicker.Stop()
			return
		}
	}
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
		return fmt.Errorf("encoder is not initialized")
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
	fw.done <- true
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
