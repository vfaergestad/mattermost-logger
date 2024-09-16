package websocket

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io"
	"mattermost-message-monitor/config"
	"mattermost-message-monitor/utils"
	"os"
)

type Client struct {
	cfg     *config.Config
	log     *zap.Logger
	conn    *websocket.Conn
	encoder *json.Encoder
	writer  *bufio.Writer
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

	// Open file for appending messages
	file, err := os.OpenFile(cfg.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Error opening output file", zap.String("File", cfg.OutputFile), zap.Error(err))
	}

	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)

	return &Client{
		cfg:     cfg,
		log:     log,
		conn:    conn,
		encoder: encoder,
		writer:  writer,
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
		HandleMessage(messageBytes, c.cfg, c.log, c.encoder, c.writer)
	}
}

func (c *Client) Close() {
	err := c.writer.Flush()
	if err != nil {
		c.log.Error("Flush error", zap.Error(err))
		return
	}
	err = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		c.log.Error("Write close error", zap.Error(err))
		return
	}
	err = c.conn.Close()
	if err != nil {
		c.log.Error("Close error", zap.Error(err))
		return
	}
}
