package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var username = flag.String("username", "TestUser", "username for the chat")

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type JoinChannelPayload struct {
	ChannelID string `json:"channel_id"`
	Username  string `json:"username"`
}

type SendMessagePayload struct {
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	q := u.Query()
	q.Set("username", *username)
	u.RawQuery = q.Encode()

	log.Printf("Connecting to %s as %s", u.String(), *username)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Start reading messages from server
	go func() {
		defer close(interrupt)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			
			var wsMsg WebSocketMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				log.Printf("unmarshal error: %v", err)
				continue
			}
			
			handleServerMessage(wsMsg)
		}
	}()

	// Join general channel by default
	joinChannel(c, "general")

	// Print usage instructions
	fmt.Println("\n=== Discord-like WebSocket Client ===")
	fmt.Println("Commands:")
	fmt.Println("  /join <channel>  - Join a channel")
	fmt.Println("  /channels        - List available channels")
	fmt.Println("  /users <channel> - List users in a channel")
	fmt.Println("  /help            - Show this help")
	fmt.Println("  /quit            - Exit the client")
	fmt.Println("  <message>        - Send a message to current channel")
	fmt.Println("\nType your messages and press Enter to send.")
	fmt.Printf("You are in channel: general\n\n")

	currentChannel := "general"
	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			
			// Send close message and wait for the close
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-time.After(time.Second):
			}
			return

		default:
			if scanner.Scan() {
				text := strings.TrimSpace(scanner.Text())
				if text == "" {
					continue
				}

				if strings.HasPrefix(text, "/") {
					// Handle commands
					parts := strings.Fields(text)
					command := parts[0]

					switch command {
					case "/join":
						if len(parts) > 1 {
							channelName := parts[1]
							joinChannel(c, channelName)
							currentChannel = channelName
							fmt.Printf("Switched to channel: %s\n", channelName)
						} else {
							fmt.Println("Usage: /join <channel>")
						}

					case "/channels":
						getChannels(c)

					case "/users":
						if len(parts) > 1 {
							channelName := parts[1]
							getUsers(c, channelName)
						} else {
							getUsers(c, currentChannel)
						}

					case "/help":
						fmt.Println("\nCommands:")
						fmt.Println("  /join <channel>  - Join a channel")
						fmt.Println("  /channels        - List available channels")
						fmt.Println("  /users <channel> - List users in a channel")
						fmt.Println("  /help            - Show this help")
						fmt.Println("  /quit            - Exit the client")
						fmt.Println("  <message>        - Send a message to current channel")

					case "/quit":
						interrupt <- os.Interrupt
						return

					default:
						fmt.Printf("Unknown command: %s. Type /help for available commands.\n", command)
					}
				} else {
					// Send message to current channel
					sendMessage(c, currentChannel, text)
				}
			}

			if err := scanner.Err(); err != nil {
				log.Printf("Scanner error: %v", err)
				return
			}
		}
	}
}

func handleServerMessage(msg WebSocketMessage) {
	switch msg.Type {
	case "message":
		// Parse message payload
		if payload, ok := msg.Payload.(map[string]interface{}); ok {
			username := payload["username"].(string)
			content := payload["content"].(string)
			msgType := payload["type"].(string)
			timestamp := payload["timestamp"].(string)

			// Parse timestamp
			t, err := time.Parse(time.RFC3339, timestamp)
			if err != nil {
				t = time.Now()
			}
			timeStr := t.Format("15:04:05")

			if msgType == "join" || msgType == "leave" {
				fmt.Printf("[%s] %s\n", timeStr, content)
			} else {
				fmt.Printf("[%s] %s: %s\n", timeStr, username, content)
			}
		}

	case "joined_channel":
		if payload, ok := msg.Payload.(map[string]interface{}); ok {
			message := payload["message"].(string)
			fmt.Printf("✓ %s\n", message)
		}

	case "user_list":
		if payload, ok := msg.Payload.(map[string]interface{}); ok {
			channelID := payload["channel_id"].(string)
			users := payload["users"].([]interface{})
			
			fmt.Printf("\nUsers in #%s:\n", channelID)
			for _, user := range users {
				if userMap, ok := user.(map[string]interface{}); ok {
					username := userMap["username"].(string)
					fmt.Printf("  • %s\n", username)
				}
			}
			fmt.Println()
		}

	case "channel_list":
		if payload, ok := msg.Payload.(map[string]interface{}); ok {
			channels := payload["channels"].([]interface{})
			
			fmt.Println("\nAvailable channels:")
			for _, channel := range channels {
				if channelMap, ok := channel.(map[string]interface{}); ok {
					name := channelMap["name"].(string)
					description := channelMap["description"].(string)
					fmt.Printf("  # %s - %s\n", name, description)
				}
			}
			fmt.Println()
		}
	}
}

func joinChannel(c *websocket.Conn, channelID string) {
	msg := WebSocketMessage{
		Type: "join_channel",
		Payload: JoinChannelPayload{
			ChannelID: channelID,
			Username:  *username,
		},
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	
	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("write error: %v", err)
	}
}

func sendMessage(c *websocket.Conn, channelID, content string) {
	msg := WebSocketMessage{
		Type: "send_message",
		Payload: SendMessagePayload{
			ChannelID: channelID,
			Content:   content,
		},
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	
	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("write error: %v", err)
	}
}

func getChannels(c *websocket.Conn) {
	msg := WebSocketMessage{
		Type:    "get_channels",
		Payload: map[string]interface{}{},
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	
	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("write error: %v", err)
	}
}

func getUsers(c *websocket.Conn, channelID string) {
	msg := WebSocketMessage{
		Type: "get_users",
		Payload: map[string]string{
			"channel_id": channelID,
		},
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	
	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("write error: %v", err)
	}
}