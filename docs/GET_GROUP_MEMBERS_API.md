# Group Members API Documentation

## Get Group Members (Paginated)

Retrieves the list of active members for a specific group with pagination support. This endpoint allows you to efficiently browse large member lists by fetching them in smaller chunks.

### Endpoint

- **URL**: `/api/v1/groups/{group_id}/members`
- **Method**: `GET`
- **Auth**: Required (Bearer Token)
- **Query Parameters**:
  - `limit` (optional): Maximum number of members to return. Default: 50, Maximum: 100.
  - `offset` (optional): Starting position for pagination. Default: 0.

### Response

The response returns a paginated list of active members along with pagination metadata.

#### Success Response (200 OK)

```json
{
  "members": [
    {
      "user_id": "uuid-user-1",
      "name": "John Doe",
      "photo_url": "https://example.com/john.jpg",
      "role": "ADMIN",
      "status": "ACTIVE",
      "joined_at": "2023-10-01T10:00:00Z"
    },
    {
      "user_id": "uuid-user-2",
      "name": "Jane Smith",
      "photo_url": "",
      "role": "MEMBER",
      "status": "ACTIVE",
      "joined_at": "2023-10-02T14:30:00Z"
    }
  ],
  "pagination": {
    "total_count": 25,
    "limit": 10,
    "offset": 0,
    "has_more": true,
    "next_offset": 10
  }
}
```

#### Fields Description

**Member Object**
- **user_id**: Unique identifier of the member
- **name**: Display name of the member
- **photo_url**: URL to the member's profile photo (empty string if not set)
- **role**: Member's role in the group (`ADMIN` or `MEMBER`)
- **status**: Member's status (always `ACTIVE` for this endpoint)
- **joined_at**: Timestamp when the member joined the group

**Pagination Object**
- **total_count**: Total number of active members in the group
- **limit**: Maximum number of members returned in this request
- **offset**: Starting position for this page of results
- **has_more**: Boolean indicating if more members are available
- **next_offset**: Suggested offset for the next page (current offset + number of returned members)

### Usage Examples

#### Get first 10 members
```bash
GET /api/v1/groups/group-uuid-123/members?limit=10&offset=0
```

#### Get next page of members
```bash
GET /api/v1/groups/group-uuid-123/members?limit=10&offset=10
```

#### Get all members (no pagination)
```bash
GET /api/v1/groups/group-uuid-123/members
# Returns all active members (up to default limit of 50)
```

### Error Responses

- **400 Bad Request**: Invalid parameters (e.g., missing group ID, invalid limit/offset values)
- **401 Unauthorized**: Invalid or missing authentication token
- **403 Forbidden**: User is not a member of the group
- **404 Not Found**: Group does not exist

### Notes

- Only active members are returned in the response
- Members are ordered by their join date (oldest first)
- The maximum limit is 100 members per request for performance reasons
- Use the pagination metadata to implement infinite scroll or traditional pagination UI