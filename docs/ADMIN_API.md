# Admin Invitation APIs

This document covers admin-focused APIs for inviting users to groups and checking registered contacts. All endpoints follow the event-driven contract: Endpoint → Command → Event → Projection. Endpoints are idempotent and return acknowledgments only.

## 1) Check Registered Contacts
- Method: POST
- URL: /api/v1/users/contacts/registered
- Auth: Required (Bearer)
- Purpose: Given a list of phone numbers, returns those registered on Tuno to simplify invitations.

Request Body:
{
  "phone_numbers": ["+14155550100", "+14155550101"]
}

Response 200:
{
  "registered_contacts": [
    {
      "id": "uuid-user-1",
      "phone_number": "+14155550100",
      "name": "Alice",
      "photo_url": "https://cdn.tuno/app/u1.jpg",
      "is_online": false,
      "last_seen_at": "2026-03-01T12:01:02Z"
    }
  ]
}

Errors:
- 400: Missing/too many phone_numbers
- 401: Unauthorized
- 500: System error

Notes:
- No state changes. Read-only; does not emit commands.

## 2) Add Group Member by Phone (Admin Only)
- Method: POST
- URL: /api/v1/groups/{group_id}/members/by-phone
- Auth: Required (Bearer)
- Role: Admin of the group
- Purpose: Add a user by phone number to a group. Verifies availability like WhatsApp.

Request Body:
{
  "command_id": "0a3b30b6-6d48-4f8d-9d1e-1b1b7a0d9a90",
  "phone_number": "+14155550100",
  "timestamp": 1741026000
}

Behavior:
- Validates admin membership.
- Looks up user by phone_number; rejects if not registered.
- Idempotent by command_id.
- Emits GroupMemberAddedEvent; projection persists membership and system message.

Response 201:
{
  "group_id": "uuid-group",
  "user_id": "uuid-user",
  "phone_number": "+14155550100",
  "status": "ACTIVE",
  "message": "Member added successfully"
}

Errors:
- 400: Phone not registered or user already in group
- 403: Not an admin or inactive
- 401: Unauthorized
- 500: System error

Logging (server-side):
- Logs via middleware; endpoints avoid business logic and only dispatch commands.

