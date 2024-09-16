package models

import "time"

// Post represents the structure of the 'post' field within a 'posted' event
type Post struct {
	ID            string                 `json:"id"`
	CreateAt      int64                  `json:"create_at"`
	UpdateAt      int64                  `json:"update_at"`
	EditAt        int64                  `json:"edit_at"`
	DeleteAt      int64                  `json:"delete_at"`
	IsPinned      bool                   `json:"is_pinned"`
	UserID        string                 `json:"user_id"`
	ChannelID     string                 `json:"channel_id"`
	RootID        string                 `json:"root_id"`
	OriginalID    string                 `json:"original_id"`
	Message       string                 `json:"message"`
	Type          string                 `json:"type"`
	Props         map[string]interface{} `json:"props"`
	Hashtags      string                 `json:"hashtags"`
	PendingPostID string                 `json:"pending_post_id"`
	ReplyCount    int                    `json:"reply_count"`
	LastReplyAt   int64                  `json:"last_reply_at"`
	Participants  interface{}            `json:"participants"` // Assuming it's nullable
	Metadata      map[string]interface{} `json:"metadata"`
}

// Message represents a simplified Mattermost message structure
type Message struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"` // Time when the message was created on Mattermost
	UserID      string    `json:"user_id"`
	ChannelID   string    `json:"channel_id"`
	ChannelName string    `json:"channel_name"`
	Message     string    `json:"message"`
	Username    string    `json:"username"`
	ProcessedAt time.Time `json:"processed_at"` // Time when the message was processed by the application
}
