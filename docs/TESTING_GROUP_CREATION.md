# Testing Group Creation (WhatsApp Style)

This document explains how to test the Group Creation flow, which follows the strict **Command → Event → Projection** pattern.

## Architecture Overview

1.  **Endpoint (`POST /api/v1/groups`)**:
    *   Accepts the request.
    *   Validates input.
    *   Dispatches `CreateGroupCommand`.
2.  **Command Handler (`CreateGroupHandler`)**:
    *   Validates business rules (e.g., Creator exists).
    *   Generates `group_id` and `invite_link`.
    *   **Emits `GroupCreatedEvent`** (Source of Truth).
    *   Returns acknowledgment (Group ID, Invite Link) to the user.
3.  **Event Handler (`GroupCreatedEventHandler`)**:
    *   Listens for `GroupCreatedEvent`.
    *   **Projects state**: Inserts Group and Members into the database in a single transaction.

## Prerequisites

Ensure the database schema is set up:

```sql
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    photo_url VARCHAR(255),
    contribution_amount NUMERIC(15, 2) NOT NULL,
    rotation_frequency VARCHAR(50) NOT NULL,
    custom_frequency_days INT,
    creator_id UUID NOT NULL,
    invite_link VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_members (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL REFERENCES groups(id),
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    joined_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL REFERENCES groups(id),
    sender_id UUID, -- Nullable for system messages
    content TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

Ensure the server is running:

```bash
make run
```

## Test Steps

### 1. Create Users (if not exist)

You need at least two users: a Creator and an Invitee.

**Creator:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+254700000001"}'

# Verify OTP (use the OTP returned in previous step)
curl -X POST http://localhost:8081/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+254700000001", "otp": "123456"}'
# Save the user_id from response as CREATOR_ID
```

**Invitee:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+254700000002"}'

# Verify OTP
curl -X POST http://localhost:8081/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+254700000002", "otp": "123456"}'
```

### 2. Create Group

Send a request to create a group. Replace `CREATOR_ID` with the actual ID.

```bash
curl -X POST http://localhost:8081/api/v1/groups/ \
  -H "Content-Type: application/json" \
  -H "X-User-ID: <CREATOR_ID>" \
  -d '{
  "command_id": "unique-uuid-123456",
  "timestamp": 1709241234,
  "name": "Family Savings",
  "photo_url": "https://example.com/group.jpg",
  "contribution_amount": 1000.00,
  "rotation_frequency": "WEEKLY", 
  "invitees": ["+254700000002", "+254700000099"] 
}'
```

**Expected Response:**

```json
{
  "group_id": "c1ab7ab6-...",
  "invite_link": "https://tuno.app/join/abcdef12",
  "invited_count": 1,
  "invited_users": ["+254700000002"],
  "failed_invitees": ["+254700000099"],
  "message": "Group created successfully"
}
```

### 3. Verify Database State

Check if the group and members were created.

```sql
SELECT * FROM groups;
SELECT * FROM group_members;
SELECT * FROM messages;
```

You should see:
1.  A group record with the returned `group_id`.
2.  Two member records:
    *   Creator: `role='ADMIN'`, `status='ACTIVE'`
    *   Invitee (+254700000002): `role='MEMBER'`, `status='INVITED'`
3.  A system message:
    *   `type='SYSTEM'`
    *   `content='User created the group'`

## Troubleshooting

*   **"failed to publish group created event"**: Check logs for Event Bus errors.
*   **"failed to project group state"**: Check database connection and transaction logs.
*   **Response OK but no DB records**: This shouldn't happen as the Event Bus is currently synchronous. If it does, check if the Event Handler is registered correctly in `main.go`.
