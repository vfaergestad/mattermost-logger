package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"mattermost-message-monitor/config"
	"mattermost-message-monitor/logger"
	"mattermost-message-monitor/utils"
	"mattermost-message-monitor/websocket"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		os.Exit(1)
	}

	// Ensure log directory exists
	logDir := filepath.Dir(cfg.LogFile)
	if err := utils.EnsureDir(logDir); err != nil {
		fmt.Printf("Failed to create log directory %s: %v\n", logDir, err)
		os.Exit(1)
	}

	// Initialize logger with log file
	log, err := logger.InitLogger(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Configuration loaded successfully",
		zap.String("MattermostDomain", cfg.MattermostDomain),
		zap.Bool("UseTLS", cfg.UseTLS),
		zap.Int("ChannelCount", len(cfg.ChannelIDs)),
		zap.String("OutputFile", cfg.OutputFile),
		zap.Bool("InsecureSkipTLSVerify", cfg.InsecureSkipTLSVerify),
		zap.String("LogFile", cfg.LogFile),
	)

	// Initialize WebSocket client
	wsClient, err := websocket.NewClient(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize WebSocket client", zap.Error(err))
	}
	defer wsClient.Close()

	// Start listening to WebSocket events
	go wsClient.Listen()

	// Handle graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	log.Info("Interrupt received, shutting down...")
	wsClient.Close()
}
