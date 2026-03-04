# Group Rounds Schedule API Documentation

## Get Group Rounds Schedule

Retrieves the complete schedule of rounds for a specific group. This endpoint provides details about each round, including the designated receiver (who gets the pot) and the contribution status of all members for that specific round.

It allows filtering rounds by their status (e.g., to see only future or completed rounds).

### Endpoint

- **URL**: `/api/v1/groups/{group_id}/schedule`
- **Method**: `GET`
- **Auth**: Required (Bearer Token)
- **Query Parameters**:
  - `status` (optional): Filter rounds by status. Allowed values: `PENDING`, `ACTIVE`, `COMPLETED`, `CANCELLED`.

### Response

The response returns a list of `rounds`. Each round object contains its scheduling details, the receiver's information, and a list of members with their payment status for that round.

#### Success Response (200 OK)

```json
{
  "rounds": [
    {
      "round_number": 1,
      "status": "COMPLETED",
      "start_date": "2023-10-01T00:00:00Z",
      "end_date": "2023-10-31T23:59:59Z",
      "receiver": {
        "user_id": "uuid-user-1",
        "name": "John Doe",
        "photo_url": "https://example.com/john.jpg"
      },
      "members": [
        {
          "user_id": "uuid-user-2",
          "name": "Jane Smith",
          "photo_url": "",
          "contribution_status": "PAID",
          "paid_at": "2023-10-05T14:30:00Z"
        },
        {
          "user_id": "uuid-user-3",
          "name": "Bob Brown",
          "photo_url": null,
          "contribution_status": "PAID",
          "paid_at": "2023-10-02T09:15:00Z"
        }
      ],
      "amount_collected": 1000.00,
      "amount_expected": 1000.00
    },
    {
      "round_number": 2,
      "status": "ACTIVE",
      "start_date": "2023-11-01T00:00:00Z",
      "end_date": "2023-11-30T23:59:59Z",
      "receiver": {
        "user_id": "uuid-user-2",
        "name": "Jane Smith",
        "photo_url": ""
      },
      "members": [
        {
          "user_id": "uuid-user-1",
          "name": "John Doe",
          "photo_url": "https://example.com/john.jpg",
          "contribution_status": "PAID",
          "paid_at": "2023-11-03T10:00:00Z"
        },
        {
          "user_id": "uuid-user-3",
          "name": "Bob Brown",
          "photo_url": null,
          "contribution_status": "PENDING",
          "paid_at": null
        }
      ],
      "amount_collected": 500.00,
      "amount_expected": 1000.00
    }
  ]
}
```

#### Fields Description

**Round Object**
- **round_number**: The sequential number of the round.
- **status**: Current status of the round (`PENDING`, `ACTIVE`, `COMPLETED`, `CANCELLED`).
- **start_date**: Date when contributions for this round begin.
- **end_date**: Deadline for contributions (payout date).
- **amount_collected**: Total amount collected from members so far.
- **amount_expected**: Total expected amount (Contribution Amount * Number of Members).

**Receiver Object (`receiver`)**
- **user_id**: ID of the member receiving the pot for this round.
- **name**: Display name of the receiver.
- **photo_url**: URL to the receiver's profile photo.

**Members List (`members`)**
- **user_id**: ID of the member contributing.
- **name**: Display name.
- **photo_url**: Profile photo URL.
- **contribution_status**: Status of their payment for *this specific round*.
  - `PAID`: Payment confirmed.
  - `PENDING`: Payment not yet made.
  - `FAILED`: Payment attempt failed.
  - `EXEMPT`: Member is exempt (e.g., the receiver, if applicable).
- **paid_at**: Timestamp when the payment was completed (null if unpaid).

### Error Responses

- **400 Bad Request**: Invalid parameters (e.g., missing group ID).
- **401 Unauthorized**: Invalid or missing token.
- **403 Forbidden**: User is not a member of the group.
- **404 Not Found**: Group does not exist.
