package websocket

import (
	"bufio"
	"encoding/json"
	"strings"
	"time"

	"mattermost-message-monitor/config"
	"mattermost-message-monitor/models"
	"mattermost-message-monitor/utils"

	"go.uber.org/zap"
)

func HandleMessage(messageBytes []byte, cfg *config.Config, log *zap.Logger, encoder *json.Encoder, writer *bufio.Writer) {
	var msg map[string]interface{}
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		log.Error("JSON Unmarshal error", zap.Error(err))
		return
	}

	event, ok := msg["event"].(string)
	if !ok {
		log.Warn("Event field missing or not a string")
		return
	}

	if event != "posted" {
		// Log non-'posted' events at DEBUG level
		log.Debug("Received non-'posted' event", zap.String("EventType", event))
		return
	}

	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		log.Warn("Data field missing or not a map in 'posted' event")
		return
	}

	// Extract the 'post' field, which is a JSON string
	postStr, ok := data["post"].(string)
	if !ok {
		log.Warn("Post field missing or not a string in 'posted' event")
		return
	}

	// Unmarshal the 'post' JSON string into the Post struct
	var post models.Post
	if err := json.Unmarshal([]byte(postStr), &post); err != nil {
		log.Error("Error unmarshaling 'post' field", zap.Error(err))
		return
	}

	// Check if the message is in one of the monitored channels
	if !utils.Contains(cfg.ChannelIDs, post.ChannelID) {
		log.Debug("Channel ID not monitored", zap.String("ChannelID", post.ChannelID))
		return
	}

	// Extract channel name from 'data' field
	channelName := utils.GetString(data, "channel_display_name")
	if channelName == "" {
		channelName = utils.GetString(data, "channel_name") // Fallback if 'channel_display_name' is empty
	}

	if channelName == "" {
		channelName = "Unknown Channel"
		log.Warn("Channel name missing in 'posted' event", zap.String("ChannelID", post.ChannelID))
	}

	// Convert 'create_at' from milliseconds to time.Time
	createdAt := time.Unix(0, post.CreateAt*int64(time.Millisecond))

	// Extract username from 'sender_name'
	username := utils.GetString(data, "sender_name")
	// Remove the "@" prefix if present
	username = strings.TrimPrefix(username, "@")

	// Populate the Message struct
	message := models.Message{
		ID:          post.ID,
		CreatedAt:   createdAt,
		UserID:      post.UserID,
		ChannelID:   post.ChannelID,
		ChannelName: channelName,
		Message:     post.Message,
		Username:    username,
		ProcessedAt: time.Now(),
	}

	// Log the message writing at INFO level
	log.Info("Message written to file",
		zap.String("MessageID", message.ID),
		zap.String("File", cfg.OutputFile),
		zap.String("ChannelName", message.ChannelName),
	)

	// Write the message as JSON to the file
	if err := encoder.Encode(message); err != nil {
		log.Error("JSON Encode error", zap.Error(err))
	}

	if err := writer.Flush(); err != nil {
		log.Error("Buffer flush error after encoding", zap.Error(err))
	}
}
