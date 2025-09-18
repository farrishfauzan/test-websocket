package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/farrishfauzan/test-websocket/internal/websocket"
)

// Server holds the handlers and dependencies
type Server struct {
	hub *websocket.Hub
}

// NewServer creates a new server with handlers
func NewServer(hub *websocket.Hub) *Server {
	return &Server{
		hub: hub,
	}
}

// HomeHandler serves the main page
func (s *Server) HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/static/index.html")
}

// WebSocketHandler handles WebSocket connections
func (s *Server) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket.ServeWS(s.hub, w, r)
}

// ChannelsHandler returns list of channels
func (s *Server) ChannelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channels := s.hub.GetChannels()
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	if err := json.NewEncoder(w).Encode(channels); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ChannelUsersHandler returns users in a channel
func (s *Server) ChannelUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id parameter is required", http.StatusBadRequest)
		return
	}

	users := s.hub.GetChannelUsers(channelID)
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HealthHandler returns server health status
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"message": "Discord-like WebSocket API is running",
	})
}