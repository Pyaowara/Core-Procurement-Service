# Approval Service API Documentation

## Overview

The **Approval Service** manages approval workflows for Purchase Requests (PR) using a fixed 4-step sequence.

It supports two integration modes:
- Manual creation via HTTP: `POST /approvals`
- Automatic creation via RabbitMQ event: `pr.ready.for.approval`

## Base URL

```text
http://localhost:6770
```

## Technology Stack

- Framework: Gin (Go)
- Database: PostgreSQL + GORM
- Message Broker: RabbitMQ (topic exchange)
- Port: `6770`

---

## Workflow

The service uses this fixed sequence:

```text
1. Employee
2. Manager
3. PurchaseOfficer
4. EXECUTIVE
```

Workflow status transitions:

```text
PENDING -> APPROVED (when step 4 is approved)
PENDING -> REJECTED (if any current step is rejected)
```

---

## Role-Based Access Control

The approval service enforces role-based validation on all approval/rejection endpoints. User roles are extracted from the JWT token and verified against the required role for the current approval step.

### Role Mapping

| Approval Step | Required Role | Allowed User Roles |
|---|---|---|
| Step 1 (Employee) | Employee | Employee, Manager, PurchaseOfficer, Executive, Admin |
| Step 2 (Manager) | Manager | Manager, Executive, Admin |
| Step 3 (PurchaseOfficer) | PurchaseOfficer | PurchaseOfficer, Executive, Admin |
| Step 4 (EXECUTIVE) | EXECUTIVE | Executive, Admin |

### Behavior

- Each approval/rejection endpoint validates the user's role from the JWT token before processing
- If role does not match, a `400 Bad Request` error is returned with message: `"user role 'xxx' cannot approve/reject step N (requires xxx)"`
- Only the current step can be approved or rejected
- Step 1 is auto-approved when PR is submitted, so role validation is usually only needed from step 2 onwards

---

## Data Models

### ApprovalInstance

```json
{
  "id": 1,
  "entity_type": "PR",
  "entity_id": 42,
  "workflow_id": "WF_42_20260308120530",
  "status": "PENDING",
  "current_step": 2,
  "created_by": 100,
  "steps": [],
  "actions": [],
  "created_at": "2026-03-08T12:05:31Z",
  "updated_at": "2026-03-08T12:05:31Z"
}
```

Notes:
- `workflow_id` is required and unique.
- It is intended as a correlation ID from Purchase Service.

### ApprovalStep

```json
{
  "id": 1,
  "instance_id": 1,
  "step_order": 1,
  "approver_id": 100,
  "role": "Employee",
  "status": "APPROVED",
  "action_at": "2026-03-08T12:05:31Z",
  "created_at": "2026-03-08T12:05:31Z",
  "updated_at": "2026-03-08T12:05:31Z"
}
```

Notes:
- `approver_id` defaults to `0` in role-based mode for pending steps.
- On PR submit flow, step 1 (`Employee`) is auto-approved by requester.
- Current implementation does not enforce role from JWT in service logic yet.

### ApprovalAction

```json
{
  "id": 1,
  "instance_id": 1,
  "step_id": 1,
  "actor_id": 100,
  "action_type": "APPROVED",
  "comment": "Looks good",
  "created_at": "2026-03-08T12:10:00Z"
}
```

---

## Endpoints

### 1. Health

`GET /health`

Response:

```json
{
  "status": "ok",
  "service": "approval-service"
}
```

---

### 2. Create Approval (Manual)

`POST /approvals`

Request:

```json
{
  "entity_type": "PR",
  "entity_id": 42,
  "created_by": 100,
  "workflow_id": "WF_42_20260308120530"
}
```

Required fields:
- `entity_type`
- `entity_id`
- `created_by`
- `workflow_id`

Response: `201 Created` with full instance payload.

---

### 3. Get Approval by Entity

`GET /approvals/:entity_type/:entity_id`

Example:

```text
GET /approvals/PR/42
```

Response: `200 OK` with full instance payload.

---

### 4. Approve Current Step by Instance ID

`POST /approvals/:id/approve`

Request:

```json
{
  "comment": "Approved"
}
```

Optional: the request body can be empty `{}` or omitted entirely.

**Authentication:**
- JWT token in `Authorization: Bearer <token>` or cookie
- User ID and role are automatically extracted from JWT token
- No need to specify `actor_id` in body

**Role Requirements:**
- Step 1 (Employee): Any user can approve (auto-approved on submit)
- Step 2 (Manager): Manager, Executive, or Admin role
- Step 3 (PurchaseOfficer): PurchaseOfficer, Executive, or Admin role
- Step 4 (EXECUTIVE): Executive or Admin role

Response: `200 OK` with updated instance.

**Error Responses:**
- `400 Bad Request`: User's role does not match required role for current step
- `401 Unauthorized`: No user_id or role found in JWT token

---

### 5. Reject Current Step by Instance ID

`POST /approvals/:id/reject`

Request:

```json
{
  "comment": "Rejected"
}
```

Optional: the request body can be empty `{}` or omitted entirely.

**Authentication:**
- JWT token in `Authorization: Bearer <token>` or cookie
- User ID and role are automatically extracted from JWT token
- No need to specify `actor_id` in body

**Role Requirements:**
- Step 1 (Employee): Any user can reject (rarely used-step 1 is auto-approved)
- Step 2 (Manager): Manager, Executive, or Admin role
- Step 3 (PurchaseOfficer): PurchaseOfficer, Executive, or Admin role
- Step 4 (EXECUTIVE): Executive or Admin role

Response: `200 OK` with updated instance.

**Error Responses:**
- `400 Bad Request`: User's role does not match required role for current step
- `401 Unauthorized`: No user_id or role found in JWT token

---

### 6. Get Approval by Workflow ID

`GET /approvals/workflow/:workflow_id`

Example:

```text
GET /approvals/workflow/WF_42_20260308120530
```

Response: `200 OK` with full instance payload.

---

### 7. Approve by Workflow ID

`POST /approvals/workflow/:workflow_id/approve`

Request:

```json
{
  "comment": "Approved"
}
```

Optional: the request body can be empty `{}` or omitted entirely.

**Authentication:**
- JWT token in `Authorization: Bearer <token>` or cookie
- User ID and role are automatically extracted from JWT token
- No need to specify `actor_id` in body

**Role-Based Access:**
See role requirements in endpoint 4 (Approve Current Step by Instance ID).

Response: `200 OK` with updated instance.

**Error Responses:**
- `400 Bad Request`: User's role does not match required role for current step
- `401 Unauthorized`: No user_id or role found in JWT token

---

### 8. Reject by Workflow ID

`POST /approvals/workflow/:workflow_id/reject`

Request:

```json
{
  "comment": "Rejected"
}
```

Optional: the request body can be empty `{}` or omitted entirely.

**Authentication:**
- JWT token in `Authorization: Bearer <token>` or cookie
- User ID and role are automatically extracted from JWT token
- No need to specify `actor_id` in body

**Role-Based Access:**
See role requirements in endpoint 5 (Reject Current Step by Instance ID).

Response: `200 OK` with updated instance.

**Error Responses:**
- `400 Bad Request`: User's role does not match required role for current step
- `401 Unauthorized`: No user_id or role found in JWT token

---

## Workflow Rejection Behavior

Rejecting a workflow immediately stops the approval process:

**How to Reject:**
- Use either endpoint (by ID or by workflow ID)
- Post a reject request on the **current step**

**What Happens on Reject:**
1. Current step status becomes `REJECTED`
2. Entire workflow status becomes `REJECTED` (not just current step)
3. An `approval.rejected` event is published to RabbitMQ with rejection reason
4. No further steps can be approved
5. Rejection is permanent — workflow cannot resume

**Example Flow:**
```
1. PR submitted -> Step 1 (Employee) auto-approved
2. current_step = 2 (Manager pending)
3. Reject at step 2 → entire workflow REJECTED
4. Steps 3 & 4 never evaluated
```

**Note:** Only the current step can be rejected. Past approved steps cannot be retroactively rejected.

---

## RabbitMQ Integration

### Connection and Topology

- Exchange: `procurement`
- Exchange type: `topic`
- Subscriber queue: `approval-pr-queue`
- Subscribed routing key: `pr.ready.for.approval`

### Subscribed Event (Incoming)

`pr.ready.for.approval`

Payload:

```json
{
  "pr_id": 42,
  "pr_number": "PR-2026-001",
  "requester_id": 100,
  "department": "IT",
  "workflow_id": "WF_42_20260308120530",
  "timestamp": "2026-03-08T12:05:30Z"
}
```

Behavior:
- Approval Service consumes the event.
- If no workflow exists for that PR, it auto-creates one with 4 steps.
- Step 1 (`Employee`) is auto-approved when the workflow is created from PR submit.
- Workflow starts waiting at step 2 (`Manager`).
- `workflow_id` from event is saved in `approval_instances.workflow_id`.

### Published Events (Outgoing)

1. `approval.completed`

```json
{
  "pr_id": 42,
  "workflow_id": "approval-15",
  "status": "APPROVED",
  "approved_at": "2026-03-08T12:30:00Z"
}
```

2. `approval.rejected`

```json
{
  "pr_id": 42,
  "workflow_id": "approval-15",
  "reason": "Insufficient justification",
  "rejected_at": "2026-03-08T12:20:00Z"
}
```

Note:
- Current code emits `workflow_id` in published events as `approval-{instance_id}`.

---

## End-to-End Flow with Purchase Service

1. User creates PR in Purchase Service.
2. User submits PR in Purchase Service.
3. Purchase Service publishes `pr.ready.for.approval` with `workflow_id`.
4. Approval Service auto-creates workflow from event.
5. Step 1 is already approved; client approves/rejects remaining steps using workflow endpoints.
6. Approval Service publishes `approval.completed` or `approval.rejected`.
7. Purchase Service consumes those events and updates PR status.

---

## Postman Test Guide

### Environment Variables

```json
{
  "auth_url": "http://localhost:6767",
  "purchase_url": "http://localhost:6769",
  "approval_url": "http://localhost:6770",
  "token": "",
  "user_id": "",
  "pr_id": "",
  "workflow_id": ""
}
```

### Step 1: Login

`POST {{auth_url}}/auth/login`

```json
{
  "username": "john_doe",
  "password": "password123"
}
```

Save:
- `token = response.token`
- `user_id = response.user.id`

### Step 2: Create PR

`POST {{purchase_url}}/pr`

Headers:
- `Authorization: Bearer {{token}}`

Save:
- `pr_id = response.data.id`

### Step 3: Submit PR

`POST {{purchase_url}}/pr/{{pr_id}}/submit`

Headers:
- `Authorization: Bearer {{token}}`

Save:
- `workflow_id = response.workflow_id`

Wait 2-3 seconds for async consumer.

### Step 4: Verify Workflow Created

`GET {{approval_url}}/approvals/workflow/{{workflow_id}}`

Expect:
- `status = PENDING`
- `current_step = 2`
- 4 steps exist

Also expect:
- step 1 (`Employee`) is already `APPROVED`
- first pending step is step 2 (`Manager`)

### Step 5-7: Approve Remaining 3 Steps

Repeat 3 times:

`POST {{approval_url}}/approvals/workflow/{{workflow_id}}/approve`

```json
{
  "actor_id": {{user_id}},
  "comment": "Approved"
}
```

After 3rd approval, expect:
- Approval instance `status = APPROVED`

### Step 9: Verify PR Final Status

`GET {{purchase_url}}/pr/{{pr_id}}`

Expect:
- `status = APPROVED`

---

## Error Response Format

```json
{
  "error": "message"
}
```

Common status codes:
- `200` success
- `201` created
- `400` bad request
- `404` not found
- `500` server error

---

## Environment Variables

```env
DB_DSN=host=localhost user=postgres password=password dbname=approval_db port=5432 sslmode=disable
PORT=6770
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

For Docker Compose networking, prefer:

```env
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
```
