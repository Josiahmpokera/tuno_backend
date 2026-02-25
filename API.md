# Tuno Backend API Documentation

All API endpoints are prefixed with `/api/v1`.

## Authentication Endpoints

The authentication flow requires phone number verification via OTP.

**Typical Flow:**
1.  **Request OTP**: Send POST request to `/api/v1/auth/otp` with the phone number.
2.  **Receive OTP**: For development, the OTP is returned in the response. In production, it would be sent via SMS/WhatsApp.
3.  **Register/Login**: Use the phone number and the received OTP to register or login.

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

### 2. Register User

Register a new user with their phone number, OTP, name, and optional photo URL.

- **URL**: `/api/v1/auth/register`
- **Method**: `POST`
- **Content-Type**: `application/json`

**Request Body:**

```json
{
  "phone_number": "+254700000001",
  "name": "Josi Ampokera",
  "photo_url": "https://example.com/photo.jpg",
  "otp": "123456"
}
```

- `phone_number` (string, required): The user's phone number.
- `name` (string, required): The user's full name.
- `photo_url` (string, optional): URL to the user's profile photo.
- `otp` (string, required): The 6-digit OTP received from `/api/v1/auth/otp`.

**Success Response:**

- **Code**: `201 Created`
- **Content**:

```json
{
  "user_id": "b63be9e5-0cf0-402d-be9d-e29421e5938c",
  "token": "mock-jwt-token-..."
}
```

**Error Responses:**

- **Code**: `400 Bad Request`
  - Content: `{"error": "OTP verification failed: invalid OTP"}`
  - Content: `{"error": "user with phone number ... already exists"}`

---

### 3. Login User

Login an existing user using their phone number and OTP.

- **URL**: `/api/v1/auth/login`
- **Method**: `POST`
- **Content-Type**: `application/json`

**Request Body:**

```json
{
  "phone_number": "+254700000001",
  "otp": "123456"
}
```

- `phone_number` (string, required): The user's phone number.
- `otp` (string, required): The 6-digit OTP received from `/api/v1/auth/otp`.

**Success Response:**

- **Code**: `200 OK`
- **Content**:

```json
{
  "user_id": "b63be9e5-0cf0-402d-be9d-e29421e5938c",
  "token": "mock-jwt-token-..."
}
```

**Error Responses:**

- **Code**: `400 Bad Request`
  - Content: `{"error": "OTP verification failed: invalid OTP"}`
- **Code**: `401 Unauthorized` (or 400 currently)
  - Content: `{"error": "user not found"}`

## Other Endpoints

### Root

- **URL**: `/api/v1/`
- **Method**: `GET`
- **Description**: Returns project title and description.

### Health Check

- **URL**: `/api/v1/health`
- **Method**: `GET`
- **Description**: Returns health status of the service.

### WebSocket

- **URL**: `/api/v1/ws`
- **Method**: `GET`
- **Description**: WebSocket connection endpoint.

## Notes

- **OTP Validity**: OTPs are valid for 5 minutes.
- **One-Time Use**: Once an OTP is verified successfully (for register or login), it is invalidated and cannot be used again.
- **Tokens**: The API currently returns a mock token. This will be replaced by a real JWT implementation in future updates.
