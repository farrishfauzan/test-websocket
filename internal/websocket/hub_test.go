package websocket

import (
	"testing"

	"github.com/farrishfauzan/test-websocket/internal/models"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	
	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}
	
	if len(hub.channels) != 2 {
		t.Errorf("Expected 2 default channels, got %d", len(hub.channels))
	}
	
	if _, exists := hub.channels["general"]; !exists {
		t.Error("Expected 'general' channel to exist")
	}
	
	if _, exists := hub.channels["random"]; !exists {
		t.Error("Expected 'random' channel to exist")
	}
}

func TestRegisterUser(t *testing.T) {
	hub := NewHub()
	
	user := hub.RegisterUser(nil, "testuser")
	
	if user == nil {
		t.Fatal("RegisterUser() returned nil")
	}
	
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	
	if user.ID == "" {
		t.Error("Expected user ID to be generated")
	}
	
	if user.Channels == nil {
		t.Error("Expected user.Channels to be initialized")
	}
	
	// Check if user is stored in hub
	if _, exists := hub.users[user.ID]; !exists {
		t.Error("User not stored in hub.users")
	}
}

func TestGetChannels(t *testing.T) {
	hub := NewHub()
	
	channels := hub.GetChannels()
	
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
	}
	
	// Check if channels have the expected properties
	for _, channel := range channels {
		if channel.ID == "" {
			t.Error("Channel ID should not be empty")
		}
		if channel.Name == "" {
			t.Error("Channel Name should not be empty")
		}
		if channel.CreatedAt.IsZero() {
			t.Error("Channel CreatedAt should be set")
		}
	}
}

func TestJoinChannel(t *testing.T) {
	hub := NewHub()
	
	// Register a user
	user := hub.RegisterUser(nil, "testuser")
	
	// Join a channel
	err := hub.JoinChannel(user.ID, "general")
	if err != nil {
		t.Errorf("Unexpected error joining channel: %v", err)
	}
	
	// Check if user is in the channel
	if !user.Channels["general"] {
		t.Error("User should be in 'general' channel")
	}
	
	// Check if channel has the user
	generalChannel := hub.channels["general"]
	if _, exists := generalChannel.Users[user.ID]; !exists {
		t.Error("Channel should contain the user")
	}
	
	// Check if join message was created
	if len(generalChannel.Messages) == 0 {
		t.Error("Expected at least one message (join message)")
	}
	
	joinMsg := generalChannel.Messages[len(generalChannel.Messages)-1]
	if joinMsg.Type != "join" {
		t.Errorf("Expected join message type, got '%s'", joinMsg.Type)
	}
}

func TestSendMessage(t *testing.T) {
	hub := NewHub()
	
	// Register a user and join a channel
	user := hub.RegisterUser(nil, "testuser")
	hub.JoinChannel(user.ID, "general")
	
	// Send a message
	testContent := "Hello, world!"
	err := hub.SendMessage(user.ID, "general", testContent)
	if err != nil {
		t.Errorf("Unexpected error sending message: %v", err)
	}
	
	// Check if message was stored
	generalChannel := hub.channels["general"]
	if len(generalChannel.Messages) < 2 { // join message + sent message
		t.Error("Expected at least 2 messages")
	}
	
	// Find the sent message (not the join message)
	var sentMessage *models.Message
	for _, msg := range generalChannel.Messages {
		if msg.Type == "message" && msg.Content == testContent {
			sentMessage = msg
			break
		}
	}
	
	if sentMessage == nil {
		t.Error("Sent message not found")
	} else {
		if sentMessage.UserID != user.ID {
			t.Errorf("Expected message UserID '%s', got '%s'", user.ID, sentMessage.UserID)
		}
		if sentMessage.Username != user.Username {
			t.Errorf("Expected message Username '%s', got '%s'", user.Username, sentMessage.Username)
		}
		if sentMessage.ChannelID != "general" {
			t.Errorf("Expected message ChannelID 'general', got '%s'", sentMessage.ChannelID)
		}
		if sentMessage.Content != testContent {
			t.Errorf("Expected message Content '%s', got '%s'", testContent, sentMessage.Content)
		}
	}
}

func TestGetChannelUsers(t *testing.T) {
	hub := NewHub()
	
	// Register users and join a channel
	user1 := hub.RegisterUser(nil, "user1")
	user2 := hub.RegisterUser(nil, "user2")
	
	hub.JoinChannel(user1.ID, "general")
	hub.JoinChannel(user2.ID, "general")
	
	// Get channel users
	users := hub.GetChannelUsers("general")
	
	if len(users) != 2 {
		t.Errorf("Expected 2 users in channel, got %d", len(users))
	}
	
	// Check if users are correct
	usernames := make(map[string]bool)
	for _, user := range users {
		usernames[user.Username] = true
	}
	
	if !usernames["user1"] {
		t.Error("Expected user1 to be in channel")
	}
	if !usernames["user2"] {
		t.Error("Expected user2 to be in channel")
	}
}

func TestSendMessageToNonExistentChannel(t *testing.T) {
	hub := NewHub()
	
	user := hub.RegisterUser(nil, "testuser")
	
	// Try to send message to non-existent channel
	err := hub.SendMessage(user.ID, "nonexistent", "test message")
	
	// Should not return error, but should be handled gracefully
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestSendMessageUserNotInChannel(t *testing.T) {
	hub := NewHub()
	
	user := hub.RegisterUser(nil, "testuser")
	
	// Don't join the channel, try to send message
	err := hub.SendMessage(user.ID, "general", "test message")
	
	// Should not return error, but message should not be sent
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Check that no message was added to the channel
	generalChannel := hub.channels["general"]
	if len(generalChannel.Messages) > 0 {
		t.Error("Message should not have been sent to channel user is not in")
	}
}