# Tuno Backend API Documentation

All API endpoints are prefixed with `/api/v1`.

## Authentication Flow (WhatsApp Style)

The authentication flow mimics WhatsApp's behavior:
1.  **Request OTP**: User enters phone number. System sends OTP.
2.  **Verify OTP**: User enters OTP. System verifies.
    *   If user exists: Logs in.
    *   If user is new: Creates a placeholder account and logs in.
3.  **Profile Setup**: If the user is new, the client prompts for Name/Photo and updates the profile.

---

### 1. Request OTP

Request an OTP for a phone number.

- **URL**: `/api/v1/auth/otp`
- **Method**: `POST`
- **Content-Type**: `application/json`

**Request Body:**

```json
{
  "phone_number": "+254700000001"
}
```

**Success Response:**

- **Code**: `200 OK`
- **Content**:

```json
{
  "message": "OTP sent successfully",
  "otp": "123456" // Returned for development/testing purposes
}
```

---

### 2. Verify OTP (Login / Register)

Verify the OTP. This endpoint handles both registration and login.

- **URL**: `/api/v1/auth/verify`
- **Method**: `POST`
- **Content-Type**: `application/json`

**Request Body:**

```json
{
  "phone_number": "+254700000001",
  "otp": "123456"
}
```

**Success Response:**

- **Code**: `200 OK`
- **Content**:

```json
{
  "user_id": "b63be9e5-0cf0-402d-be9d-e29421e5938c",
  "token": "mock-jwt-token-...",
  "is_new_user": true, // true if user was just created, false if existing
  "user": {
      "id": "...",
      "phone_number": "...",
      "name": "", // Empty if new user
      "photo_url": ""
  }
}
```

**Client Logic:**
- If `is_new_user` is `true`, navigate to **Profile Setup** screen.
- If `is_new_user` is `false`, navigate to **Home** screen.

---

## User Endpoints

### 3. Update Profile

Update the user's name and photo URL. Typically called after registration or from settings.

- **URL**: `/api/v1/users/profile`
- **Method**: `PUT`
- **Content-Type**: `application/json`

**Request Body:**

```json
{
  "user_id": "b63be9e5-0cf0-402d-be9d-e29421e5938c", // TODO: Will be removed once JWT middleware is active
  "name": "Josi Ampokera",
  "photo_url": "https://example.com/photo.jpg"
}
```

**Success Response:**

- **Code**: `200 OK`
- **Content**:

```json
{
  "id": "b63be9e5-0cf0-402d-be9d-e29421e5938c",
  "phone_number": "+254700000001",
  "name": "Josi Ampokera",
  "photo_url": "https://example.com/photo.jpg",
  "created_at": "...",
  "updated_at": "..."
}
```

---

## Group Endpoints

### 4. Create Group

Create a new rotational savings group and invite users.

- **URL**: `/api/v1/groups/`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Headers**:
    - `X-User-ID`: `{user_id}` (Required for now until JWT middleware is fully integrated)

**Request Body:**

```json
{
  "name": "Family Savings",
  "photo_url": "https://example.com/group.jpg",
  "contribution_amount": 1000.00,
  "rotation_frequency": "WEEKLY", // Options: EVERY_DAY, EVERY_2_DAYS, EVERY_3_DAYS, WEEKLY, MONTHLY, CUSTOM
  "custom_frequency_days": 0, // Only if rotation_frequency is CUSTOM
  "invitees": ["+254700000002", "+254700000003"] // List of phone numbers to invite
}
```

**Success Response:**

- **Code**: `201 Created`
- **Content**:

```json
{
  "group_id": "c1ab7ab6-2496-453f-aa5b-34ec6409eefa",
  "invite_link": "https://tuno.app/join/abcdef12",
  "invited_count": 2,
  "invited_users": ["+254700000002", "+254700000003"],
  "failed_invitees": [],
  "message": "Group created successfully"
}
```

---

## Other Endpoints

### Root

- **URL**: `/api/v1/`
- **Method**: `GET`
- **Description**: Returns API version and info.

### Health Check

- **URL**: `/api/v1/health`
- **Method**: `GET`
- **Description**: Returns system health status.

### WebSocket

- **URL**: `/api/v1/ws`
- **Method**: `GET`
- **Description**: WebSocket endpoint for real-time communication.
