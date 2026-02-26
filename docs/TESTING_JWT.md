# Testing JWT Authentication

This document describes how to test the newly implemented JWT authentication flow.

## 1. Prerequisites

Ensure the server is running:

```bash
go run cmd/api/main.go
```

Ensure Redis and PostgreSQL are running.

## 2. Authentication Flow

The flow consists of two steps:
1.  **Request OTP**: Send a phone number to receive an OTP.
2.  **Verify OTP**: Submit the OTP to receive a JWT token.

### Step 1: Request OTP

**Endpoint:** `POST /api/v1/auth/otp`

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/otp \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+254700000000"
  }'
```

**Response:**

```json
{
  "message": "OTP sent successfully",
  "otp": "123456" 
}
```

*Note: For development, the OTP is returned in the response.*

### Step 2: Verify OTP & Get Token

**Endpoint:** `POST /api/v1/auth/verify`

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+254700000000",
    "otp": "123456"
  }'
```

**Response:**

```json
{
  "is_new_user": false,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user-uuid",
    "phone_number": "+254700000000",
    ...
  },
  "user_id": "user-uuid"
}
```

**Copy the `token` from the response.** You will need it for authenticated requests.

## 3. accessing Protected Routes

The Group Creation endpoint is now protected and requires a Bearer token.

### Create Group

**Endpoint:** `POST /api/v1/groups/`

**Headers:**
- `Authorization: Bearer <YOUR_JWT_TOKEN>`
- `Content-Type: application/json`

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/groups/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -d '{
    "command_id": "unique-uuid-for-idempotency",
    "timestamp": 1709241234,
    "name": "My Savings Group",
    "contribution_amount": 500,
    "rotation_frequency": "WEEKLY",
    "invitees": ["+254711111111", "+254722222222"]
  }'
```

**Response (Success):**

```json
{
  "group_id": "generated-group-uuid",
  "status": "created"
}
```

**Response (Unauthorized - Missing Token):**

```json
{
  "error": "Authorization header is required"
}
```

**Response (Unauthorized - Invalid Token):**

```json
{
  "error": "Invalid or expired token"
}
```

## 4. Verification Checklist

- [ ] Can request OTP?
- [ ] Can verify OTP and receive a JWT token?
- [ ] Can create a group WITH a valid token?
- [ ] Is group creation REJECTED without a token?
- [ ] Is the correct User ID used from the token (check logs/database)?
