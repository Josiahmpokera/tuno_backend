# Testing WebSocket with Postman

This guide explains how to test the Tuno WebSocket API using Postman.

## **Important: Use a WebSocket Request**
Postman has a dedicated mode for WebSockets. **Do not** use a standard HTTP GET request.

1. In Postman, click **New** (top left).
2. Select **WebSocket Request**.
3. In the URL bar, enter: `ws://localhost:8081/api/v1/ws`

---

## **Method 1: Query Parameter (Easiest)**
This method passes the token directly in the URL.

1. **URL**: `ws://localhost:8081/api/v1/ws`
2. **Params Tab**:
   - Key: `token`
   - Value: `<your_jwt_token>`
   *(Postman will update the URL to `ws://localhost:8081/api/v1/ws?token=...`)*
3. Click **Connect**.
4. You should see "Connected" in green.

---

## **Method 2: Authorization Header (More Secure)**
This method passes the token in the HTTP headers during the handshake.

1. **URL**: `ws://localhost:8081/api/v1/ws`
2. **Headers Tab**:
   - Key: `Authorization`
   - Value: `Bearer <your_jwt_token>`
   *(Make sure to include the word "Bearer " followed by a space)*
3. Click **Connect**.
4. You should see "Connected" in green.

---

## **Sending a Message**
Once connected, you can send messages to a group.

1. **Prerequisite**: You need a valid `group_id`. Use `GET /api/v1/groups` to find one.
2. In the **Message** area (bottom panel), change the format dropdown from "Text" to **JSON**.
3. Paste the message payload:
   ```json
   {
     "group_id": "YOUR_GROUP_ID_HERE",
     "content": "Hello from Postman!",
     "type": "TEXT"
   }
   ```
4. Click **Send**.
5. You should receive a response in the **Messages** pane below (both your sent message echo and messages from others).

---

## **Alternative: Sending via REST API**
You can also send messages using a standard HTTP POST request. This is useful for mobile apps or external integrations.

1. **Method**: `POST`
2. **URL**: `http://localhost:8081/api/v1/groups/:id/messages` (Replace `:id` with actual Group ID)
3. **Headers**:
   - `Authorization`: `Bearer <token>`
   - `Content-Type`: `application/json`
4. **Body (JSON)**:
   ```json
   {
     "content": "Hello via REST!",
     "type": "TEXT"
   }
   ```
5. **Response**: 201 Created with message details.
   - Note: Users connected via WebSocket will receive this message in real-time.
