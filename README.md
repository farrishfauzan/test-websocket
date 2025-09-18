# Discord-like WebSocket API in Go

A real-time chat application built with Go and WebSockets, featuring Discord-like functionality including channels, users, and message broadcasting.

## Features

- 🔌 **Real-time WebSocket Communication** - Instant messaging with WebSocket connections
- 📢 **Channel-based Chat** - Support for multiple chat channels/rooms
- 👥 **User Management** - Track online users and their presence in channels
- 💬 **Message Broadcasting** - Efficient message distribution to channel members
- 🌟 **Discord-like UI** - Modern web interface similar to Discord
- 🔧 **REST API Endpoints** - HTTP endpoints for channel and user information
- 📱 **Responsive Design** - Works on desktop and mobile devices
- 🚀 **Easy to Deploy** - Single binary deployment with embedded static files

## Quick Start

### Prerequisites

- Go 1.19 or higher
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/farrishfauzan/test-websocket.git
cd test-websocket
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the server:
```bash
go run cmd/server/main.go
```

4. Open your browser and navigate to `http://localhost:8080`

### Using Custom Port

```bash
go run cmd/server/main.go -addr=:3000
```

## API Endpoints

### WebSocket Endpoint
- **WS** `/ws?username=<name>` - WebSocket connection for real-time chat

### REST API Endpoints
- **GET** `/health` - Health check endpoint
- **GET** `/api/channels` - List all available channels
- **GET** `/api/channels/users?channel_id=<id>` - List users in a specific channel

### WebSocket Message Types

#### Client to Server Messages

1. **Join Channel**
```json
{
  "type": "join_channel",
  "payload": {
    "channel_id": "general",
    "username": "user123"
  }
}
```

2. **Send Message**
```json
{
  "type": "send_message",
  "payload": {
    "channel_id": "general",
    "content": "Hello everyone!"
  }
}
```

3. **Get Channels**
```json
{
  "type": "get_channels",
  "payload": {}
}
```

4. **Get Users**
```json
{
  "type": "get_users",
  "payload": {
    "channel_id": "general"
  }
}
```

#### Server to Client Messages

1. **Message**
```json
{
  "type": "message",
  "payload": {
    "id": "uuid",
    "user_id": "user-uuid",
    "username": "user123",
    "channel_id": "general",
    "content": "Hello everyone!",
    "timestamp": "2023-01-01T12:00:00Z",
    "type": "message"
  }
}
```

2. **User List**
```json
{
  "type": "user_list",
  "payload": {
    "channel_id": "general",
    "users": [
      {
        "id": "user-uuid",
        "username": "user123"
      }
    ]
  }
}
```

## Example Usage

### Web Interface

1. Open `http://localhost:8080` in your browser
2. Enter a username in the sidebar
3. Click on channels to switch between them
4. Type messages and press Enter to send

### Command Line Client

Run the example CLI client:

```bash
go run examples/client.go -username=TestUser
```

Available commands:
- `/join <channel>` - Join a channel
- `/channels` - List available channels
- `/users <channel>` - List users in a channel
- `/help` - Show help
- `/quit` - Exit client

### cURL Examples

```bash
# Health check
curl http://localhost:8080/health

# Get all channels
curl http://localhost:8080/api/channels

# Get users in a channel
curl "http://localhost:8080/api/channels/users?channel_id=general"
```

## Architecture

### Project Structure

```
├── cmd/
│   └── server/          # Main server application
├── internal/
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   └── websocket/       # WebSocket hub and client logic
├── web/
│   └── static/          # Static web files
├── examples/            # Example client implementations
└── docs/               # Documentation
```

### Components

- **Hub**: Central coordinator for WebSocket connections and message broadcasting
- **Client**: Manages individual WebSocket connections and message handling
- **Models**: Data structures for users, channels, and messages
- **Handlers**: HTTP request handlers for REST API endpoints

## Default Channels

The server starts with two default channels:
- **general** - General discussion channel
- **random** - Random chat channel

## Development

### Building

```bash
go build -o server cmd/server/main.go
```

### Running Tests

```bash
go test ./...
```

### Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/web ./web
EXPOSE 8080
CMD ["./server"]
```

Build and run:
```bash
docker build -t websocket-chat .
docker run -p 8080:8080 websocket-chat
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Gorilla WebSocket](https://github.com/gorilla/websocket)
- Inspired by Discord's user interface and functionality
- Uses [Google UUID](https://github.com/google/uuid) for unique identifiers