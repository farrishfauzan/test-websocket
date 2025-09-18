# API Documentation

This document describes the REST API and WebSocket API for the Discord-like WebSocket chat application.

## REST API Endpoints

### Health Check

**GET** `/health`

Returns the server health status.

**Response:**
```json
{
  "status": "ok",
  "message": "Discord-like WebSocket API is running"
}
```

### Get Channels

**GET** `/api/channels`

Returns a list of all available channels.

**Response:**
```json
[
  {
    "id": "general",
    "name": "General",
    "description": "General discussion channel",
    "messages": [],
    "created_at": "2023-01-01T12:00:00Z"
  },
  {
    "id": "random",
    "name": "Random",
    "description": "Random chat channel",
    "messages": [],
    "created_at": "2023-01-01T12:00:00Z"
  }
]
```

### Get Channel Users

**GET** `/api/channels/users?channel_id=<channel_id>`

Returns a list of users currently in the specified channel.

**Parameters:**
- `channel_id` (required): The ID of the channel

**Response:**
```json
[
  {
    "id": "user-uuid-1",
    "username": "user1"
  },
  {
    "id": "user-uuid-2",
    "username": "user2"
  }
]
```

## WebSocket API

### Connection

**WebSocket Endpoint:** `/ws?username=<username>`

**Parameters:**
- `username` (optional): Username for the connection. Defaults to "Anonymous" if not provided.

### Message Format

All WebSocket messages follow this format:

```json
{
  "type": "message_type",
  "payload": {
    // type-specific payload
  }
}
```

## Client to Server Messages

### Join Channel

Join a specific channel to start receiving messages from it.

```json
{
  "type": "join_channel",
  "payload": {
    "channel_id": "general",
    "username": "user123"
  }
}
```

### Send Message

Send a message to a channel you've joined.

```json
{
  "type": "send_message",
  "payload": {
    "channel_id": "general",
    "content": "Hello everyone!"
  }
}
```

### Get Channels

Request a list of all available channels.

```json
{
  "type": "get_channels",
  "payload": {}
}
```

### Get Users

Request a list of users in a specific channel.

```json
{
  "type": "get_users",
  "payload": {
    "channel_id": "general"
  }
}
```

## Server to Client Messages

### Message

Represents a chat message, join/leave notification, or system message.

```json
{
  "type": "message",
  "payload": {
    "id": "message-uuid",
    "user_id": "user-uuid",
    "username": "user123",
    "channel_id": "general",
    "content": "Hello everyone!",
    "timestamp": "2023-01-01T12:00:00Z",
    "type": "message"
  }
}
```

**Message Types:**
- `"message"`: Regular chat message
- `"join"`: User joined the channel
- `"leave"`: User left the channel
- `"system"`: System message

### Joined Channel

Confirmation that you've successfully joined a channel.

```json
{
  "type": "joined_channel",
  "payload": {
    "channel_id": "general",
    "message": "Successfully joined channel"
  }
}
```

### Channel List

List of all available channels in response to `get_channels`.

```json
{
  "type": "channel_list",
  "payload": {
    "channels": [
      {
        "id": "general",
        "name": "General",
        "description": "General discussion channel",
        "messages": [],
        "created_at": "2023-01-01T12:00:00Z"
      }
    ]
  }
}
```

### User List

List of users in a channel in response to `get_users` or when users join/leave.

```json
{
  "type": "user_list",
  "payload": {
    "channel_id": "general",
    "users": [
      {
        "id": "user-uuid-1",
        "username": "user1"
      },
      {
        "id": "user-uuid-2",
        "username": "user2"
      }
    ]
  }
}
```

## Example Usage

### JavaScript WebSocket Client

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?username=TestUser');

ws.onopen = function() {
    console.log('Connected to WebSocket');
    
    // Join a channel
    ws.send(JSON.stringify({
        type: 'join_channel',
        payload: {
            channel_id: 'general',
            username: 'TestUser'
        }
    }));
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
    
    if (data.type === 'message') {
        console.log(`${data.payload.username}: ${data.payload.content}`);
    }
};

// Send a message
function sendMessage(content) {
    ws.send(JSON.stringify({
        type: 'send_message',
        payload: {
            channel_id: 'general',
            content: content
        }
    }));
}
```

### cURL Examples

```bash
# Health check
curl http://localhost:8080/health

# Get all channels
curl http://localhost:8080/api/channels

# Get users in general channel
curl "http://localhost:8080/api/channels/users?channel_id=general"
```

## Error Handling

The WebSocket API handles errors gracefully:

- Unknown message types are logged but don't cause disconnection
- Invalid channel IDs are ignored silently
- Users not in channels cannot send messages to those channels
- Connection drops are handled with automatic cleanup

## Rate Limiting

Currently, there is no rate limiting implemented. In a production environment, you should implement:

- Message rate limiting per user
- Connection rate limiting per IP
- Maximum message size limits (currently 512 bytes)

## Security Considerations

- CORS is enabled for all origins (development only)
- No authentication is implemented
- Usernames are not validated for uniqueness
- Messages are not persisted beyond server runtime

For production use, consider implementing:
- User authentication and authorization
- Message persistence
- Input validation and sanitization
- Rate limiting
- HTTPS/WSS encryption
- CORS restrictions