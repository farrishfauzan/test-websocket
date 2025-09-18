package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/farrishfauzan/test-websocket/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Channels in the server
	channels map[string]*models.Channel

	// Users in the server
	users map[string]*models.User

	// Mutex for thread safety
	mutex sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	hub := &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		channels:   make(map[string]*models.Channel),
		users:      make(map[string]*models.User),
	}

	// Create default general channel
	hub.createDefaultChannels()
	
	return hub
}

// createDefaultChannels creates some default channels
func (h *Hub) createDefaultChannels() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// General channel
	generalChannel := &models.Channel{
		ID:          "general",
		Name:        "General",
		Description: "General discussion channel",
		Users:       make(map[string]*models.User),
		Messages:    make([]*models.Message, 0),
		CreatedAt:   time.Now(),
	}
	h.channels["general"] = generalChannel

	// Random channel
	randomChannel := &models.Channel{
		ID:          "random",
		Name:        "Random",
		Description: "Random chat channel",
		Users:       make(map[string]*models.User),
		Messages:    make([]*models.Message, 0),
		CreatedAt:   time.Now(),
	}
	h.channels["random"] = randomChannel
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client registered: %s", client.user.Username)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.handleUserDisconnect(client)
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: %s", client.user.Username)
			}

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// handleUserDisconnect handles user disconnection
func (h *Hub) handleUserDisconnect(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	user := client.user
	if user == nil {
		return
	}

	// Remove user from all channels
	for channelID := range user.Channels {
		if channel, exists := h.channels[channelID]; exists {
			delete(channel.Users, user.ID)

			// Send leave message to channel
			leaveMsg := &models.Message{
				ID:        uuid.New().String(),
				UserID:    user.ID,
				Username:  user.Username,
				ChannelID: channelID,
				Content:   user.Username + " left the channel",
				Timestamp: time.Now(),
				Type:      "leave",
			}
			channel.Messages = append(channel.Messages, leaveMsg)

			// Broadcast leave message to channel users
			h.broadcastToChannel(channelID, &models.WebSocketMessage{
				Type:    "message",
				Payload: leaveMsg,
			})

			// Broadcast updated user list
			h.broadcastUserList(channelID)
		}
	}

	// Remove user from users map
	delete(h.users, user.ID)
}

// broadcastMessage broadcasts a message to all clients
func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// broadcastToChannel broadcasts a message to all users in a specific channel
func (h *Hub) broadcastToChannel(channelID string, message *models.WebSocketMessage) {
	h.mutex.RLock()
	channel, exists := h.channels[channelID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for _, user := range channel.Users {
		if user.Conn != nil {
			select {
			case <-time.After(10 * time.Second):
				log.Printf("Timeout writing to user %s", user.Username)
				continue
			default:
				err := user.Conn.WriteMessage(websocket.TextMessage, messageBytes)
				if err != nil {
					log.Printf("Error writing to user %s: %v", user.Username, err)
				}
			}
		}
	}
}

// JoinChannel adds a user to a channel
func (h *Hub) JoinChannel(userID, channelID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	user, userExists := h.users[userID]
	if !userExists {
		return nil // User doesn't exist
	}

	channel, channelExists := h.channels[channelID]
	if !channelExists {
		return nil // Channel doesn't exist
	}

	// Add user to channel
	channel.Users[userID] = user
	user.Channels[channelID] = true

	// Create join message
	joinMsg := &models.Message{
		ID:        uuid.New().String(),
		UserID:    userID,
		Username:  user.Username,
		ChannelID: channelID,
		Content:   user.Username + " joined the channel",
		Timestamp: time.Now(),
		Type:      "join",
	}
	channel.Messages = append(channel.Messages, joinMsg)

	// Broadcast join message to channel
	go h.broadcastToChannel(channelID, &models.WebSocketMessage{
		Type:    "message",
		Payload: joinMsg,
	})

	// Broadcast updated user list
	go h.broadcastUserList(channelID)

	return nil
}

// SendMessage sends a message to a channel
func (h *Hub) SendMessage(userID, channelID, content string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	user, userExists := h.users[userID]
	if !userExists {
		return nil
	}

	channel, channelExists := h.channels[channelID]
	if !channelExists {
		return nil
	}

	// Check if user is in the channel
	if _, inChannel := user.Channels[channelID]; !inChannel {
		return nil
	}

	// Create message
	message := &models.Message{
		ID:        uuid.New().String(),
		UserID:    userID,
		Username:  user.Username,
		ChannelID: channelID,
		Content:   content,
		Timestamp: time.Now(),
		Type:      "message",
	}
	channel.Messages = append(channel.Messages, message)

	// Broadcast message to channel
	go h.broadcastToChannel(channelID, &models.WebSocketMessage{
		Type:    "message",
		Payload: message,
	})

	return nil
}

// GetChannels returns all channels
func (h *Hub) GetChannels() []*models.Channel {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	channels := make([]*models.Channel, 0, len(h.channels))
	for _, channel := range h.channels {
		// Create a copy without the Users map to avoid circular references
		channelCopy := &models.Channel{
			ID:          channel.ID,
			Name:        channel.Name,
			Description: channel.Description,
			Messages:    channel.Messages,
			CreatedAt:   channel.CreatedAt,
		}
		channels = append(channels, channelCopy)
	}
	return channels
}

// GetChannelUsers returns users in a channel
func (h *Hub) GetChannelUsers(channelID string) []*models.User {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	channel, exists := h.channels[channelID]
	if !exists {
		return nil
	}

	users := make([]*models.User, 0, len(channel.Users))
	for _, user := range channel.Users {
		// Create a copy without the Conn to avoid exposing it
		userCopy := &models.User{
			ID:       user.ID,
			Username: user.Username,
		}
		users = append(users, userCopy)
	}
	return users
}

// broadcastUserList broadcasts the updated user list for a channel
func (h *Hub) broadcastUserList(channelID string) {
	users := h.GetChannelUsers(channelID)
	userListMsg := &models.WebSocketMessage{
		Type: "user_list",
		Payload: &models.UserListResponse{
			ChannelID: channelID,
			Users:     users,
		},
	}
	h.broadcastToChannel(channelID, userListMsg)
}

// RegisterUser registers a new user
func (h *Hub) RegisterUser(conn *websocket.Conn, username string) *models.User {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	user := &models.User{
		ID:       uuid.New().String(),
		Username: username,
		Conn:     conn,
		Channels: make(map[string]bool),
	}
	h.users[user.ID] = user
	return user
}