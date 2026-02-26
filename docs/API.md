
### Get Group Details
`GET /api/v1/groups/:id`

Retrieves detailed information about a specific group, including member count and current round status.

**Headers**
- `Authorization: Bearer <token>` (required)

**Response (200 OK)**

```json
{
  "id": "uuid",
  "name": "Group Name",
  "role": "ADMIN",
  "members_count": 5,
  "current_round": {
    "round_number": 1,
    "status": "ACTIVE",
    "recipient_id": "uuid",
    "end_date": "2023-10-01T00:00:00Z"
  }
}
```

### Get Group Members
`GET /api/v1/groups/:id/members`

Retrieves the list of members for a specific group.

**Headers**
- `Authorization: Bearer <token>` (required)

**Response (200 OK)**

```json
[
  {
    "user_id": "uuid",
    "name": "John Doe",
    "photo_url": "http://example.com/photo.jpg",
    "role": "MEMBER",
    "status": "ACTIVE",
    "joined_at": "2023-10-01T00:00:00Z"
  }
]
```

### Get Group Messages
`GET /api/v1/groups/:id/messages`

Retrieves paginated message history for a group.

**Query Parameters**
- `limit` (optional): Number of messages to return (default: 50).
- `offset` (optional): Number of messages to skip (default: 0).

**Headers**
- `Authorization: Bearer <token>` (required)

**Response (200 OK)**

```json
[
  {
    "id": "uuid",
    "group_id": "uuid",
    "sender_id": "uuid",
    "content": "Hello world",
    "type": "TEXT",
    "sequence": 1,
    "created_at": "2023-10-01T12:00:00Z"
  }
]
```

### Send Group Message
`POST /api/v1/groups/:id/messages`

Sends a message to a group. This message will be broadcast to all connected WebSocket clients in the group.

**Headers**
- `Authorization: Bearer <token>` (required)
- `Content-Type: application/json`

**Request Body**

```json
{
  "content": "Hello everyone!",
  "type": "TEXT"
}
```

**Response (201 Created)**

```json
{
  "Message": {
    "id": "uuid",
    "group_id": "uuid",
    "sender_id": "uuid",
    "content": "Hello everyone!",
    "type": "TEXT",
    "sequence": 2,
    "created_at": "2023-10-01T12:05:00Z"
  },
  "MemberIDs": ["uuid1", "uuid2"]
}
```

## Messaging

### WebSocket Connection
`GET /api/v1/ws`

Establishes a WebSocket connection for real-time messaging.

**Authentication**
You can authenticate using either a query parameter (recommended for browsers) or an Authorization header.

1. **Query Parameter:** `?token=<jwt_token>`
2. **Header:** `Authorization: Bearer <jwt_token>`

**Client Actions**
- Send Message: `{"group_id": "uuid", "content": "Hello", "type": "TEXT"}`

**Server Events**
- Receive Message: `{"id": "uuid", "group_id": "uuid", "sender_id": "uuid", "content": "Hello", "type": "TEXT", "sequence": 1, "created_at": "timestamp"}`

## Private Messaging (One-to-One, WhatsApp-Style)

### WebSocket Events

To make the conversation feel "live", the backend pushes events to all participants (sender and recipient).

**Event: `direct_message`**

When a message is sent, both users receive this payload via WebSocket:

```json
{
  "type": "direct_message",
  "id": "uuid",
  "conversation_id": "uuid",
  "sender_id": "uuid",
  "content": "Hello!",
  "message_type": "TEXT",
  "created_at": "2023-10-01T12:00:00Z"
}
```

**Frontend Strategy:**
1. **Optimistic UI:** When User A sends a message, immediately show it in the UI as "Sending...".
2. **Server Confirmation:** When the WebSocket event arrives (or HTTP response returns), update status to "Sent".
3. **Live Updates:** When User B replies, the WebSocket event will instantly show the message on User A's screen.
4. **Deduplication:** Since the sender also receives the WebSocket event, ensure your frontend checks if the message ID already exists to avoid duplicates.

### Start Conversation
`POST /api/v1/conversations`

Starts (or reuses) a private conversation between the authenticated user and another user. This only succeeds if both users share at least one common active group.

**Headers**
- `Authorization: Bearer <token>` (required)
- `Content-Type: application/json`

**Request Body**

```json
{
  "command_id": "8d3f5b3a-0a7f-4e9f-9e1b-123456789abc",
  "recipient_id": "uuid-of-recipient"
}
```

**Notes**
- `command_id` is required for idempotency. Re-sending the same `command_id` will not create duplicate conversations.
- If a conversation already exists between the two users, the existing `conversation_id` is returned.
- The backend verifies that both users still share at least one active group at the time of the request.

**Response (200 OK)**

```json
{
  "conversation_id": "uuid",
  "message": "Conversation started successfully"
}
```

**Error Responses**

```json
// 400 Bad Request (no shared active group)
{
  "error": "users do not share a common active group"
}

// 400 Bad Request (self conversation)
{
  "error": "cannot start conversation with self"
}
```

### List Conversations
`GET /api/v1/conversations`

Returns all private conversations for the authenticated user, ordered by last activity.

**Headers**
- `Authorization: Bearer <token>` (required)

**Response (200 OK)**

```json
{
  "conversations": [
    {
      "id": "uuid",
      "user1_id": "uuid",
      "user2_id": "uuid",
      "created_at": "2023-10-01T12:00:00Z",
      "updated_at": "2023-10-02T09:15:00Z"
    }
  ]
}
```

### Send Direct Message
`POST /api/v1/conversations/:id/messages`

Sends a direct message inside a private conversation. The backend re-checks that the sender and recipient still share at least one active group before accepting the message.

**Headers**
- `Authorization: Bearer <token>` (required)
- `Content-Type: application/json`

**Path Parameters**
- `id`: Conversation ID

**Request Body**

```json
{
  "command_id": "bd4b4e4b-1e1c-4f1e-8f03-abcdefabcdef",
  "content": "Hello, how are you?",
  "type": "TEXT"
}
```

**Notes**
- If the users no longer share an active group, the message is rejected.
- Old messages remain stored even if the shared group relationship breaks; you just cannot send new messages.

**Response (202 Accepted)**

```json
{
  "message": "Message sent"
}
```

**Error Responses**

```json
// 400 Bad Request (no shared active group anymore)
{
  "error": "cannot send message: no common active group shared with recipient"
}
```

### List Direct Messages
`GET /api/v1/conversations/:id/messages`

Retrieves paginated direct messages for a given conversation.

**Headers**
- `Authorization: Bearer <token>` (required)

**Query Parameters**
- `limit` (optional): Number of messages to return (default: 50).
- `offset` (optional): Number of messages to skip (default: 0).

**Response (200 OK)**

```json
{
  "messages": [
    {
      "id": "uuid",
      "conversation_id": "uuid",
      "sender_id": "uuid",
      "content": "Hello, how are you?",
      "type": "TEXT",
      "sequence_number": 1,
      "created_at": "2023-10-02T09:15:00Z"
    }
  ]
}
```
