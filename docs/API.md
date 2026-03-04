# Tuno Backend API Documentation

This documentation covers all available endpoints for the Tuno application, including Authentication, Group Management, Private Messaging (WhatsApp-style), and Real-time WebSockets.

## Base URL
`http://localhost:8080/api/v1` (Local Development)

## Authentication
All protected endpoints require a valid JWT token in the Authorization header.
Header format: `Authorization: Bearer <your_token>`

---

## 1. Authentication Module

### Send OTP
`POST /auth/otp`

Initiates the login/registration process by sending an OTP to the provided phone number.

**Request Body**
```json
{
  "phone_number": "+1234567890"
}
```

**Response (200 OK)**
```json
{
  "message": "OTP sent successfully"
}
```

### Verify OTP & Login
`POST /auth/verify`

Verifies the OTP. If valid, it returns a JWT token. If the user does not exist, they are automatically registered with an empty profile.
Check `is_new_user` in the response. If `true`, you should redirect the user to the Registration screen to complete their profile (Name & Photo).

**Request Body**
```json
{
  "phone_number": "+1234567890",
  "otp": "123456"
}
```

**Response (200 OK)**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "is_new_user": true,
  "user": {
    "id": "uuid",
    "phone_number": "+1234567890",
    "name": "",
    "photo_url": ""
  }
}
```

### Register (Complete Profile)
`POST /auth/register`

Completes the registration process by setting the user's name and profile picture. This should be called immediately after `Verify OTP` if `is_new_user` is true.

**Headers**
`Authorization: Bearer <token>`

**Request Body**
```json
{
  "name": "John Doe",
  "photo_url": "https://example.com/photo.jpg"
}
```

**Response (200 OK)**
```json
{
  "id": "uuid",
  "name": "John Doe",
  "photo_url": "https://example.com/photo.jpg"
}
```

---

## 2. User Module

### Get My Profile
`GET /users/me`

Returns the authenticated user's profile, similar to WhatsApp's personal info view.

**Headers**
`Authorization: Bearer <token>`

**Response (200 OK)**
```json
{
  "id": "uuid",
  "phone_number": "+1234567890",
  "name": "John Doe",
  "photo_url": "https://example.com/photo.jpg",
  "is_online": true,
  "last_seen_at": "2026-03-04T10:00:00Z",
  "created_at": "2026-03-01T10:00:00Z",
  "updated_at": "2026-03-04T10:05:00Z"
}
```

### Update Profile
`PUT /users/profile`

Updates the authenticated user's profile information. (Same logic as Register)

**Request Body**
```json
{
  "name": "John Doe",
  "photo_url": "https://example.com/photo.jpg"
}
```
*(Note: `user_id` is inferred from the token)*

**Response (200 OK)**
```json
{
  "id": "uuid",
  "name": "John Doe",
  "photo_url": "https://example.com/photo.jpg"
}
```

---

## 3. Group Module

### Create Group
`POST /groups`

Creates a new savings group.

**Request Body**
```json
{
  "name": "My Savings Group",
  "admin_id": "uuid-of-admin",
  "member_ids": ["uuid-1", "uuid-2"],
  "amount": 100.00,
  "currency": "USD",
  "frequency": "MONTHLY",
  "start_date": "2023-11-01T00:00:00Z"
}
```

**Response (201 Created)**
```json
{
  "group_id": "uuid",
  "message": "Group created successfully"
}
```

### List User Groups
`GET /groups`

Returns all groups the authenticated user belongs to.

**Response (200 OK)**
```json
[
  {
    "id": "uuid",
    "name": "My Savings Group",
    "role": "ADMIN", 
    "member_count": 5
  }
]
```

### Get Group Details
`GET /groups/:id`

Retrieves detailed information about a specific group.

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
`GET /groups/:id/members`

Retrieves the list of members for a specific group.

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
`GET /groups/:id/messages`

Retrieves paginated message history for a group.

**Query Parameters**
- `limit`: Number of messages (default: 50)
- `offset`: Number of messages to skip (default: 0)

**Response (200 OK)**
```json
[
  {
    "id": "uuid",
    "group_id": "uuid",
    "sender_id": "uuid",
    "content": "Hello world",
    "type": "TEXT",
    "created_at": "2023-10-01T12:00:00Z"
  }
]
```

### Send Group Message
`POST /groups/:id/messages`

Sends a message to a group.

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
    "content": "Hello everyone!",
    ...
  },
  "MemberIDs": ["uuid1", "uuid2"]
}
```

---

## 4. Private Messaging (WhatsApp-Style)

This module handles 1-on-1 conversations. It enforces strict validation: **Users must share at least one active group** to communicate.

### Start/Get Conversation
`POST /conversations`

Starts (or reuses) a private conversation between the authenticated user and another user.
**Validation:** Fails if users do not share a common active group.

**Request Body**
```json
{
  "command_id": "uuid-for-idempotency",
  "recipient_id": "uuid-of-recipient"
}
```

**Response (200 OK)**
```json
{
  "conversation_id": "uuid",
  "message": "Conversation started successfully"
}
```

### List Conversations
`GET /conversations`

Returns all private conversations for the authenticated user, ordered by last activity.

**Response (200 OK)**
```json
{
  "conversations": [
    {
      "id": "uuid",
      "user1_id": "uuid",
      "user2_id": "uuid",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

### Send Direct Message
`POST /conversations/:id/messages`

Sends a DM. The server performs a **Soft Blocking** check: if the users no longer share a group, the message is rejected.
However, old messages remain readable.

**Request Body**
```json
{
  "command_id": "uuid-for-idempotency",
  "content": "Hello!",
  "type": "TEXT"
}
```

**Response (202 Accepted)**
*The server accepts the message, persists it, and delivers it asynchronously via WebSocket.*
```json
{
  "message": "Message sent"
}
```

### Get Direct Messages
`GET /conversations/:id/messages`

Retrieves paginated DMs. Messages include a **monotonic sequence number** per conversation to help with ordering and sync.

**Response (200 OK)**
```json
{
  "messages": [
    {
      "id": "uuid",
      "sender_id": "uuid",
      "content": "Hello!",
      "sequence_number": 1,
      "status": "SENT",
      "created_at": "..."
    }
  ]
}
```

### Mark Messages as Read
`POST /conversations/:id/read`

Marks all messages in the conversation (sent by the partner) as **READ**. This triggers a `messages_read` WebSocket event to the partner (blue ticks) and the sender (for multi-device sync).

**Request Body**
```json
{
  "command_id": "uuid-for-idempotency"
}
```

**Response (200 OK)**
```json
{
  "message": "Messages marked as read"
}
```

---

## 5. Real-time WebSocket

### Connection
`GET /ws`

Establishes a WebSocket connection for real-time updates.
**Presence Tracking:** The server automatically tracks the user's **Online/Offline** status and **Last Seen** timestamp based on this connection.

**Authentication**
- Query Param: `?token=<jwt_token>` (Recommended)
- Header: `Authorization: Bearer <jwt_token>`

### Events

#### `direct_message`
Received when a DM is sent. Sent to **both** the Recipient and the Sender (for multi-device sync and confirmation).

```json
{
  "type": "direct_message",
  "id": "uuid",
  "conversation_id": "uuid",
  "sender_id": "uuid",
  "content": "Hello!",
  "message_type": "TEXT",
  "created_at": "timestamp",
  "sequence_number": 123
}
```

#### `messages_read`
Received when a partner reads messages in a conversation. Sent to **both** the Partner (Client A - The Sender) and the Reader (Client B - The Reader, for multi-device sync).

```json
{
  "type": "messages_read",
  "conversation_id": "uuid",
  "reader_id": "uuid-of-who-read-it",
  "read_at": "timestamp"
}
```

### Frontend Strategy (Optimistic UI)
1.  **Send:** User sends message -> UI shows it immediately (grey/pending).
2.  **Ack:** HTTP 202 returns -> UI keeps it pending.
3.  **Confirm:** WebSocket `direct_message` event arrives -> UI updates status to "Sent" (tick).
4.  **Receive:** Recipient receives `direct_message` event -> UI appends it to chat.
