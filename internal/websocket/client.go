package websocket

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/farrishfauzan/test-websocket/internal/models"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for development
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// User associated with this client
	user *models.User
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		
		// Handle the WebSocket message
		c.handleMessage(message)
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(message []byte) {
	var wsMsg models.WebSocketMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	switch wsMsg.Type {
	case "join_channel":
		c.handleJoinChannel(wsMsg.Payload)
	case "send_message":
		c.handleSendMessage(wsMsg.Payload)
	case "get_channels":
		c.handleGetChannels()
	case "get_users":
		c.handleGetUsers(wsMsg.Payload)
	default:
		log.Printf("Unknown message type: %s", wsMsg.Type)
	}
}

// handleJoinChannel handles joining a channel
func (c *Client) handleJoinChannel(payload interface{}) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling join channel payload: %v", err)
		return
	}

	var joinPayload models.JoinChannelPayload
	if err := json.Unmarshal(payloadBytes, &joinPayload); err != nil {
		log.Printf("Error unmarshaling join channel payload: %v", err)
		return
	}

	if c.user == nil {
		log.Printf("User not found for client")
		return
	}

	err = c.hub.JoinChannel(c.user.ID, joinPayload.ChannelID)
	if err != nil {
		log.Printf("Error joining channel: %v", err)
		return
	}

	// Send confirmation to client
	response := &models.WebSocketMessage{
		Type: "joined_channel",
		Payload: map[string]string{
			"channel_id": joinPayload.ChannelID,
			"message":    "Successfully joined channel",
		},
	}
	c.sendMessage(response)
}

// handleSendMessage handles sending a message
func (c *Client) handleSendMessage(payload interface{}) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling send message payload: %v", err)
		return
	}

	var sendPayload models.SendMessagePayload
	if err := json.Unmarshal(payloadBytes, &sendPayload); err != nil {
		log.Printf("Error unmarshaling send message payload: %v", err)
		return
	}

	if c.user == nil {
		log.Printf("User not found for client")
		return
	}

	err = c.hub.SendMessage(c.user.ID, sendPayload.ChannelID, sendPayload.Content)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// handleGetChannels handles getting channel list
func (c *Client) handleGetChannels() {
	channels := c.hub.GetChannels()
	response := &models.WebSocketMessage{
		Type: "channel_list",
		Payload: &models.ChannelListResponse{
			Channels: channels,
		},
	}
	c.sendMessage(response)
}

// handleGetUsers handles getting user list for a channel
func (c *Client) handleGetUsers(payload interface{}) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling get users payload: %v", err)
		return
	}

	var getUsersPayload map[string]string
	if err := json.Unmarshal(payloadBytes, &getUsersPayload); err != nil {
		log.Printf("Error unmarshaling get users payload: %v", err)
		return
	}

	channelID := getUsersPayload["channel_id"]
	users := c.hub.GetChannelUsers(channelID)
	response := &models.WebSocketMessage{
		Type: "user_list",
		Payload: &models.UserListResponse{
			ChannelID: channelID,
			Users:     users,
		},
	}
	c.sendMessage(response)
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(message *models.WebSocketMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case c.send <- messageBytes:
	default:
		close(c.send)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWS handles websocket requests from the peer.
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Get username from query parameter
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}

	// Register user
	user := hub.RegisterUser(conn, username)

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		user: user,
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}