# WhatsApp-Style Group Creation Logic (Go)

Follow this logic exactly.
This is **command → event → projection**.
REST or WebSocket is **only a transport pipe**.

---

## 1. Accept a Create Group Command (Pipe)

* Accept request via REST **or** WebSocket
* Do NOT write to DB here

```go
CreateGroupCommand{
  CommandID,
  CreatorID,
  GroupName,
  Members,
  Timestamp,
}
```

---

## 2. Validate Command (No Side Effects)

* Creator exists
* Members exist
* Creator allowed to create group
* Idempotency check using `CommandID`

❌ No DB writes
❌ No messages sent

---

## 3. Generate Group ID (Server Only)

* Generate globally unique `group_id`
* This ID never changes

---

## 4. Emit Immutable Event (SOURCE OF TRUTH)

```go
GroupCreatedEvent{
  GroupID,
  CreatorID,
  Members,
  CreatedAt,
}
```

👉 **This event = group exists**

---

## 5. Project State From Event (Side Effects)

From `GroupCreatedEvent`:

* Insert group record
* Insert group members
* Assign admin role to creator
* Initialize round configuration
* Insert system message:

  > “User created the group”

All inside **one transaction**.

---

## 6. Acknowledge Creator (ACK)

* Return `group_id`
* ACK only after event is committed
* Safe for retries

---

## 7. Notify Participants (AFTER Commit)

* Online → WebSocket
* Offline → Push notification
* Never part of creation logic

---

## 8. Rules (DO NOT BREAK)

* ❌ Endpoints must not mutate state
* ❌ Clients must not generate group IDs
* ❌ No business logic in handlers
* ❌ No system messages from clients
* ✅ Events create reality
* ✅ Everything else reacts

---

## 9. One-Line Mental Model

> **Accept intent → validate → record fact → let system react**

---

## 10. Testing

For detailed testing instructions, see [TESTING_GROUP_CREATION.md](TESTING_GROUP_CREATION.md).

---

END
