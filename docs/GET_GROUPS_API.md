# Group List API Documentation

## Get User Groups

Retrieves a list of groups the authenticated user is a member of. The list is ordered by the most recent activity (latest message timestamp or group creation time), similar to WhatsApp.

### Endpoint

- **URL**: `/api/v1/groups/`
- **Method**: `GET`
- **Auth**: Required (Bearer Token)

### Response

The response is a JSON array of group objects. Each object includes the user's role in the group and the last message sent in that group (if any).

#### Success Response (200 OK)

```json
[
  {
    "id": "uuid-group-1",
    "name": "Family Savings",
    "photo_url": "https://example.com/photo.jpg",
    "contribution_amount": 100.00,
    "rotation_frequency": "MONTHLY",
    "custom_frequency_days": 0,
    "creator_id": "uuid-user-1",
    "invite_link": "https://tuno.app/invite/abc",
    "created_at": "2023-10-01T10:00:00Z",
    "updated_at": "2023-10-01T10:00:00Z",
    "role": "ADMIN",
    "last_message": {
      "id": "uuid-msg-1",
      "group_id": "uuid-group-1",
      "sender_id": "uuid-user-2",
      "content": "Payment confirmed!",
      "type": "TEXT",
      "created_at": "2023-10-05T14:30:00Z"
    }
  },
  {
    "id": "uuid-group-2",
    "name": "Office Lunch",
    "photo_url": "",
    "contribution_amount": 0,
    "rotation_frequency": "WEEKLY",
    "custom_frequency_days": 0,
    "creator_id": "uuid-user-3",
    "invite_link": "https://tuno.app/invite/xyz",
    "created_at": "2023-09-15T09:00:00Z",
    "updated_at": "2023-09-15T09:00:00Z",
    "role": "MEMBER",
    "last_message": null
  }
]
```

#### Fields Description

- **id**: Unique identifier for the group.
- **name**: Name of the group.
- **photo_url**: URL to the group's profile photo.
- **contribution_amount**: Amount to be contributed per round.
- **rotation_frequency**: Frequency of contributions (e.g., "WEEKLY", "MONTHLY").
- **role**: The authenticated user's role in the group ("ADMIN" or "MEMBER").
- **last_message**: Object containing details of the last message sent in the group. `null` if no messages exist.
  - **content**: The text content of the message.
  - **type**: The type of message ("TEXT", "SYSTEM", "IMAGE").
  - **sender_id**: ID of the user who sent the message (null for system messages).
  - **created_at**: Timestamp when the message was sent.

### Error Responses

- **401 Unauthorized**: If the user is not logged in or the token is invalid.
- **500 Internal Server Error**: If there is a server-side error processing the request.
