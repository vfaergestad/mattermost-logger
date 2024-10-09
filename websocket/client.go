package websocket

import (
	"crypto/tls"
	"io"
	"mattermost-message-monitor/config"
	"mattermost-message-monitor/filewriter"
	"mattermost-message-monitor/utils"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	cfg        *config.Config
	log        *zap.Logger
	conn       *websocket.Conn
	fileWriter *filewriter.FileWriter
}

func NewClient(cfg *config.Config, log *zap.Logger) (*Client, error) {
	wsURL, err := utils.ConstructWebSocketURL(cfg.MattermostDomain, cfg.MattermostPort, cfg.UseTLS)
	if err != nil {
		log.Fatal("Error constructing WebSocket URL", zap.Error(err))
	}

	dialer := websocket.DefaultDialer
	if cfg.InsecureSkipTLSVerify { // Disable tls-verification if disabled in config
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		log.Warn("TLS certificate verification is disabled. This is insecure and should not be used in production.")
	}

	header := make(map[string][]string)
	header["Authorization"] = []string{"Bearer " + cfg.AuthToken}

	conn, resp, err := dialer.Dial(wsURL, header)
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			log.Fatal("WebSocket dial error",
				zap.Error(err),
				zap.Int("StatusCode", resp.StatusCode),
				zap.String("ResponseBody", string(body)),
			)
		}
		log.Fatal("WebSocket dial error", zap.Error(err))
	}

	log.Info("Connected to Mattermost WebSocket", zap.String("URL", wsURL))

	// Initialize FileWriter
	fw, err := filewriter.NewFileWriter(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize FileWriter", zap.Error(err))
	}

	return &Client{
		cfg:        cfg,
		log:        log,
		conn:       conn,
		fileWriter: fw,
	}, nil
}

func (c *Client) Listen() {
	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			c.log.Error("Read error", zap.Error(err))
			return
		}

		// Delegate handling to handler.go
		HandleMessage(messageBytes, c.cfg, c.log, c.fileWriter)
	}
}

func (c *Client) Close() {
	// Close the FileWriter
	if err := c.fileWriter.Close(); err != nil {
		c.log.Error("Failed to close FileWriter", zap.Error(err))
	}

	// Close the WebSocket connection
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		c.log.Error("Write close error", zap.Error(err))
	}
	err = c.conn.Close()
	if err != nil {
		c.log.Error("Close error", zap.Error(err))
	}
}
