package models

import (
	"time"
	"github.com/gorilla/websocket"
)

// User represents a connected user
type User struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	Conn     *websocket.Conn `json:"-"`
	Channels map[string]bool `json:"-"` // channels the user is in
}

// Channel represents a chat channel/room
type Channel struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Users       map[string]*User `json:"-"`
	Messages    []*Message       `json:"messages"`
	CreatedAt   time.Time        `json:"created_at"`
}

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	ChannelID string    `json:"channel_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "message", "join", "leave", "system"
}

// WebSocketMessage represents the structure of messages sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// JoinChannelPayload represents the payload for joining a channel
type JoinChannelPayload struct {
	ChannelID string `json:"channel_id"`
	Username  string `json:"username"`
}

// SendMessagePayload represents the payload for sending a message
type SendMessagePayload struct {
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
}

// ChannelListResponse represents the response for channel list
type ChannelListResponse struct {
	Channels []*Channel `json:"channels"`
}

// UserListResponse represents the response for user list in a channel
type UserListResponse struct {
	ChannelID string  `json:"channel_id"`
	Users     []*User `json:"users"`
}