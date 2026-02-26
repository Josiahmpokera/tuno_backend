# Testing Get User Groups

This document describes how to test the `GET /api/v1/groups` endpoint.

## 1. Prerequisites

Ensure the server is running:

```bash
go run cmd/api/main.go
```

Ensure you have a valid JWT token (refer to `TESTING_JWT.md`).

## 2. Get User Groups

**Endpoint:** `GET /api/v1/groups`

**Headers:**
- `Authorization: Bearer <YOUR_JWT_TOKEN>`

**Request:**

```bash
curl -X GET http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>"
```

**Expected Response (200 OK):**

```json
[
  {
    "id": "...",
    "name": "Family Savings",
    "photo_url": "...",
    "contribution_amount": 1000,
    "rotation_frequency": "WEEKLY",
    "custom_frequency_days": 0,
    "creator_id": "...",
    "invite_link": "...",
    "created_at": "...",
    "updated_at": "...",
    "role": "ADMIN"
  }
]
```

**Scenario: User has no groups**
Response should be an empty array `[]`.

**Scenario: Missing Token**
Response should be `401 Unauthorized`.
