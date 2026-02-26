# WhatsApp-Style Messaging Architecture

This document outlines the implementation of the WhatsApp-style messaging system for Tuno.

## 1. Overview

The system uses a hybrid approach:
- **REST API**: For fetching state (Group Details, Message History).
- **WebSocket**: For real-time intent (Sending messages) and events (Receiving messages).
- **Server-Side Ordering**: The server assigns sequence numbers and timestamps, acting as the single source of truth.

## 2. REST Endpoints (Read-Only)

### 2.1 Get Group Details
Use this to render the group screen header (Name, Members, Round Status).

- **Endpoint**: `GET /api/v1/groups/:id`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
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

### 2.2 Get Message History
Use this to render the chat history. Messages are ordered by `created_at` DESC (newest first).

- **Endpoint**: `GET /api/v1/groups/:id/messages?limit=50&offset=0`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
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

## 3. WebSocket Protocol (Real-Time)

### 3.1 Connection
- **URL**: `ws://localhost:8080/api/v1/ws?token=<jwt_token>`
- **Authentication**: JWT Token in query parameter.

### 3.2 Sending a Message
Client sends a JSON payload to the WebSocket connection.

- **Payload**:
```json
{
  "group_id": "uuid",
  "content": "Hello from WebSocket",
  "type": "TEXT"
}
```

### 3.3 Receiving a Message
Server broadcasts the persisted message to all group members (including sender).

- **Payload**:
```json
{
  "id": "uuid",
  "group_id": "uuid",
  "sender_id": "uuid",
  "content": "Hello from WebSocket",
  "type": "TEXT",
  "sequence": 2,
  "created_at": "2023-10-01T12:01:00Z"
}
```

## 4. Testing Guide

### Prerequisites
1. **Create User & Login** to get JWT Token.
2. **Create Group** to get Group ID.

### Step 1: Connect to WebSocket
Use `wscat` (install via `npm install -g wscat`):

```bash
wscat -c "ws://localhost:8080/api/v1/ws?token=<YOUR_TOKEN>"
```

### Step 2: Send Message
Paste this into the `wscat` terminal:

```json
{"group_id": "<GROUP_ID>", "content": "Hello Tuno", "type": "TEXT"}
```

You should receive back the message object with an ID and timestamp.

### Step 3: Verify via REST
Check if the message is persisted.

```bash
curl -X GET http://localhost:8080/api/v1/groups/<GROUP_ID>/messages \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```
