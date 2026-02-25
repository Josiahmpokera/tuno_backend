# RUles to follow
- Maintain good project folder structure
- Maintain the Exception Handling and Error Handling please
- Maintain the code quality and readability
- Maintain the code comments
- Follow Meta standart please
- Maintain the code versioning please
- Save the project for me please, Every changes Save it


# Trae Endpoint Writing Rules

**Project:** Rotational Group Savings & Messaging Platform (Upatu)

This document defines **mandatory rules** Cursor must follow when implementing API endpoints.

These rules exist to enforce:

* Event-driven architecture
* WhatsApp-style behavior
* Financial correctness
* Long-term scalability

---

## 1. Endpoint Philosophy (MANDATORY)

Every API endpoint MUST follow this pattern:

> **Endpoint = Command Intake (PIPE)**
> **Events = State Creation**
> **Handlers = State Projection**

An endpoint must NEVER:

* Directly write business state
* Contain financial logic
* Calculate rounds
* Emit WebSocket messages directly

---

## 2. Endpoint Responsibilities (ALLOWED)

An endpoint is allowed to:

1. Authenticate the user
2. Validate input format
3. Authorize access (group membership / role)
4. Create a **Command**
5. Dispatch the command
6. Return an acknowledgment

Nothing else.

---

## 3. What an Endpoint MUST NOT Do

❌ Insert/update/delete core tables
❌ Calculate round dates
❌ Determine payout recipients
❌ Emit system messages
❌ Push notifications
❌ Contain transaction logic

---

## 4. Mandatory Endpoint Structure

Every endpoint must follow this exact structure:

```go
func Handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Authenticate
    userID := ctx.Value("user_id")

    // 2. Parse input
    var req RequestDTO
    decode(r.Body, &req)

    // 3. Validate input
    validate(req)

    // 4. Create command
    cmd := DomainCommand{
        ...
    }

    // 5. Dispatch command
    err := commandBus.Dispatch(ctx, cmd)

    // 6. Respond
    respond(w, result)
}
```

Deviation is NOT allowed.

---

## 5. Naming Rules

### Endpoints

* Use **verbs sparingly**
* Focus on intent, not state

✅ `/groups`
❌ `/create-group-now`

### Commands

* Must start with a verb

Examples:

* `CreateGroupCommand`
* `ConfirmContributionCommand`
* `ConfirmPayoutCommand`

---

## 6. Idempotency Rule (CRITICAL)

All endpoints that **change state** MUST be idempotent.

### Rule

* Client may retry same request
* Server must not create duplicates

### Required Fields

* `command_id`
* `user_id`
* `timestamp`

---

## 7. Event Emission Rule

Endpoints MUST NOT emit events directly.

Only **Command Handlers** may emit events.

```text
Endpoint → Command → Event → Projection
```

---

## 8. Validation Rule

### Endpoint Validates

* Required fields
* Field formats
* Basic authorization

### Command Handler Validates

* Business rules
* Domain invariants
* Financial correctness

---

## 9. System Messages Rule

System messages:

* Are created ONLY by event handlers
* Are NEVER sent by clients
* Are NEVER created in endpoints

---

## 10. Financial Action Rules

Any endpoint that affects money (even confirmations):

* MUST create a command
* MUST emit immutable events
* MUST be auditable
* MUST be reversible only by new events

---

## 11. Read vs Write Separation

### Write Endpoints

* Emit commands
* Return acknowledgment only

### Read Endpoints

* Read projected state
* NEVER emit commands or events

---

## 12. WebSocket Rule

Endpoints must NEVER:

* Access WebSocket connections
* Push real-time messages

WebSocket fan-out is triggered ONLY by events.

---

## 13. Error Handling Rule

* Domain errors → 400 / 403
* System errors → 500
* Never leak internal details
* Never partially commit state

---

## 14. Logging Rule

Every endpoint must log:

* command_id
* user_id
* endpoint name
* result

---

## 15. Example (Correct)

```go
POST /groups

✔ Validates request
✔ Creates CreateGroupCommand
✔ Dispatches command
✔ Returns group_id
```

---

## 16. Example (FORBIDDEN)

```go
POST /groups

❌ Inserts into groups table
❌ Inserts members
❌ Sends system message
❌ Pushes WebSocket message
```

---

## 17. Final Instruction to Cursor

Cursor must:

* Follow this structure exactly
* Reject shortcuts
* Prefer clarity over cleverness
* Ask for clarification instead of guessing

This system is **event-driven**, not CRUD-driven.

---

END OF ENDPOINT WRITING RULES

