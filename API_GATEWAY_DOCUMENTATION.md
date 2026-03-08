# API Gateway Documentation

## Overview

The **API Gateway** serves as a unified entry point for all microservices in the Core Procurement System. Instead of calling multiple services on different ports, clients can route all requests through the gateway on **port 8080**.

### Why Use the Gateway?

- **Single Entry Point**: Frontend only needs to know about one URL (`localhost:8080`)
- **Simplified Routing**: All service prefixes are transparent via `/api/` namespace
- **CORS Enabled**: Cross-origin requests handled at gateway level
- **Future-Ready**: Centralized place for authentication, rate limiting, logging

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  API Gateway (Port 8080)                     │
│                   All requests start with /api/              │
└────────────┬────────────┬────────────┬────────────┬──────────┘
             │            │            │            │
     ┌───────▼──┐  ┌──────▼──┐ ┌──────▼──┐ ┌──────▼──┐
     │   Auth   │  │Inventory│ │Purchase │ │Approval │
     │  (6767)  │  │ (6768)  │ │ (6769)  │ │ (6770)  │
     └──────────┘  └─────────┘ └─────────┘ └─────────┘
```

---

## Base URL

```
http://localhost:8080/api
```

All requests must start with `/api/` prefix and include appropriate service path.

---

## Authentication

All endpoints (except `POST /api/auth/login`) require a **JWT token** in the `Authorization` header.

### Get JWT Token

1. First, login to get a token:

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

2. Response includes `token`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 100,
    "name": "John Doe",
    "email": "user@example.com",
    "role": "Employee"
  }
}
```

3. Include token in subsequent requests:

```http
Authorization: Bearer <token>
```

### User Roles

- **Employee**: Can create PRs, view own PRs
- **Manager**: Approves PRs at department level
- **PurchaseOfficer**: Approves POs, manages vendors
- **Executive**: Final approval on high-value POs
- **Admin**: Full system access

---

## Service Routing

### 1. Auth & Identity Service

**Direct Port**: 6767  
**Gateway Path**: `/api/auth/` or `/api/users/`

#### Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/auth/login` | User login | ❌ |
| POST | `/api/auth/register` | Register new user | ❌ |
| GET | `/api/users/profile` | Get current user profile | ✅ |
| GET | `/api/users/:id` | Get user by ID | ✅ |
| PUT | `/api/users/:id` | Update user | ✅ |

#### Postman Examples

##### Login

```http
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{
  "email": "john@company.com",
  "password": "SecurePassword123"
}
```

**Response**: `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 100,
    "name": "John Doe",
    "email": "john@company.com",
    "role": "Employee"
  }
}
```

##### Get Current User Profile

```http
GET http://localhost:8080/api/users/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**: `200 OK`
```json
{
  "id": 100,
  "name": "John Doe",
  "email": "john@company.com",
  "role": "Employee",
  "department": "Engineering",
  "created_at": "2026-01-15T10:30:00Z"
}
```

---

### 2. Inventory Service

**Direct Port**: 6768  
**Gateway Path**: `/api/inventory/` or `/api/dep/`

#### Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/inventory/` | List all inventory items | ✅ |
| GET | `/api/inventory/:id` | Get inventory by ID | ✅ |
| GET | `/api/inventory/health` | Service health check | ❌ |
| GET | `/api/dep/` | List departments | ✅ |

#### Postman Examples

##### List Inventory

```http
GET http://localhost:8080/api/inventory/
Authorization: Bearer <token>
```

**Response**: `200 OK`
```json
[
  {
    "id": 1,
    "product_name": "Laptop Dell XPS",
    "sku": "LAPTOP-XPS-001",
    "category": "Electronics",
    "quantity_available": 50,
    "quantity_reserved": 10,
    "unit_price": 1200.00,
    "last_updated": "2026-03-05T14:22:00Z"
  },
  {
    "id": 2,
    "product_name": "Office Chair",
    "sku": "CHAIR-OFF-001",
    "category": "Furniture",
    "quantity_available": 200,
    "quantity_reserved": 5,
    "unit_price": 250.00,
    "last_updated": "2026-03-06T09:15:00Z"
  }
]
```

##### Get Specific Item

```http
GET http://localhost:8080/api/inventory/1
Authorization: Bearer <token>
```

**Response**: `200 OK`
```json
{
  "id": 1,
  "product_name": "Laptop Dell XPS",
  "sku": "LAPTOP-XPS-001",
  "category": "Electronics",
  "quantity_available": 50,
  "quantity_reserved": 10,
  "unit_price": 1200.00,
  "last_updated": "2026-03-05T14:22:00Z"
}
```

##### Service Health Check

```http
GET http://localhost:8080/api/inventory/health
```

**Response**: `200 OK`
```json
{
  "status": "healthy",
  "service": "inventory-service",
  "timestamp": "2026-03-08T12:05:31Z"
}
```

---

### 3. Purchase Service

**Direct Port**: 6769  
**Gateway Path**: `/api/purchase/`

#### Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/purchase/health` | Service health check | ❌ |
| **Purchase Requests (PR)** | | | |
| GET | `/api/purchase/pr` | List all PRs | ✅ |
| POST | `/api/purchase/pr` | Create new PR | ✅ |
| GET | `/api/purchase/pr/:id` | Get PR details | ✅ |
| PUT | `/api/purchase/pr/:id` | Update PR | ✅ |
| DELETE | `/api/purchase/pr/:id` | Soft delete PR | ✅ |
| **Purchase Orders (PO)** | | | |
| GET | `/api/purchase/po` | List all POs | ✅ |
| POST | `/api/purchase/po` | Create PO from PR | ✅ |
| GET | `/api/purchase/po/:id` | Get PO details | ✅ |
| PUT | `/api/purchase/po/:id` | Update PO details | ✅ |
| **Vendors** | | | |
| GET | `/api/purchase/vendor` | List all vendors | ✅ |
| POST | `/api/purchase/vendor` | Create new vendor | ✅ |
| GET | `/api/purchase/vendor/:id` | Get vendor details | ✅ |

#### Postman Examples

##### Create Purchase Request (PR)

```http
POST http://localhost:8080/api/purchase/pr
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Office Supplies Q2 2026",
  "description": "Quarterly office supply purchase",
  "department": "Administration",
  "budget": 5000.00,
  "items": [
    {
      "product_name": "Printer Paper A4",
      "quantity": 100,
      "unit_price": 5.00,
      "description": "Standard white paper, 80gsm"
    },
    {
      "product_name": "Pen Set",
      "quantity": 50,
      "unit_price": 2.50,
      "description": "Ballpoint pens, assorted colors"
    }
  ]
}
```

**Response**: `201 Created`
```json
{
  "id": 1,
  "title": "Office Supplies Q2 2026",
  "description": "Quarterly office supply purchase",
  "department": "Administration",
  "budget": 5000.00,
  "total_amount": 750.00,
  "status": "PENDING_APPROVAL",
  "created_by": 100,
  "created_at": "2026-03-08T12:05:31Z",
  "updated_at": "2026-03-08T12:05:31Z"
}
```

> **Note**: After PR creation, approval workflow automatically triggers via RabbitMQ. Status progresses to `APPROVED` after all approval steps complete.

##### Get PR Details

```http
GET http://localhost:8080/api/purchase/pr/1
Authorization: Bearer <token>
```

**Response**: `200 OK`
```json
{
  "id": 1,
  "title": "Office Supplies Q2 2026",
  "description": "Quarterly office supply purchase",
  "department": "Administration",
  "budget": 5000.00,
  "total_amount": 750.00,
  "status": "APPROVED",
  "created_by": 100,
  "created_at": "2026-03-08T12:05:31Z",
  "updated_at": "2026-03-08T12:10:15Z"
}
```

##### Create Purchase Order (PO)

```http
POST http://localhost:8080/api/purchase/po
Authorization: Bearer <token>
Content-Type: application/json

{
  "pr_id": 1,
  "vendor_id": 5,
  "po_number": "PO-2026-001",
  "expected_delivery": "2026-04-15",
  "terms": "Net 30"
}
```

**Response**: `201 Created`
```json
{
  "id": 1,
  "pr_id": 1,
  "vendor_id": 5,
  "po_number": "PO-2026-001",
  "status": "PENDING_APPROVAL",
  "total_amount": 750.00,
  "expected_delivery": "2026-04-15",
  "created_at": "2026-03-08T12:05:31Z",
  "updated_at": "2026-03-08T12:05:31Z"
}
```

> **Note**: PO also triggers approval workflow automatically if configured. Check with PurchaseOfficer role.

##### List All Purchase Requests

```http
GET http://localhost:8080/api/purchase/pr
Authorization: Bearer <token>
```

**Response**: `200 OK`
```json
[
  {
    "id": 1,
    "title": "Office Supplies Q2 2026",
    "department": "Administration",
    "budget": 5000.00,
    "total_amount": 750.00,
    "status": "APPROVED"
  },
  {
    "id": 2,
    "title": "IT Equipment Budget",
    "department": "Engineering",
    "budget": 25000.00,
    "total_amount": 15000.00,
    "status": "PENDING_APPROVAL"
  }
]
```

---

### 4. Approval Service

**Direct Port**: 6770  
**Gateway Path**: `/api/approval/`

#### Overview

The Approval Service implements a 4-step role-based workflow:

1. **Employee** (PR Creator) - Automatically approved on submission
2. **Manager** (Department Head) - Reviews for departmental alignment
3. **PurchaseOfficer** - Reviews for vendor/budget compliance
4. **Executive** - Final approval for high-value items

#### Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/approval/:entity_type/:entity_id` | Get approval by entity | ✅ |
| POST | `/api/approval/:id/approve` | Approve by instance ID | ✅ |
| POST | `/api/approval/:id/reject` | Reject by instance ID | ✅ |
| GET | `/api/approval/workflows/:workflow_id` | Get workflow status | ✅ |
| POST | `/api/approval/workflows/:workflow_id/approve` | Approve by workflow ID | ✅ |
| POST | `/api/approval/workflows/:workflow_id/reject` | Reject by workflow ID | ✅ |

#### Postman Examples

##### Get Approval for PR

```http
GET http://localhost:8080/api/approval/PR/1
Authorization: Bearer <token>
```

**Response**: `200 OK`
```json
{
  "id": 1,
  "entity_type": "PR",
  "entity_id": 1,
  "workflow_id": "WF_1_20260308120530",
  "status": "IN_PROGRESS",
  "current_step": 2,
  "created_by": 100,
  "created_at": "2026-03-08T12:05:31Z",
  "steps": [
    {
      "id": 1,
      "step_order": 1,
      "role": "Employee",
      "status": "APPROVED",
      "approver_id": 100,
      "action_at": "2026-03-08T12:05:31Z"
    },
    {
      "id": 2,
      "step_order": 2,
      "role": "Manager",
      "status": "PENDING",
      "approver_id": 0,
      "action_at": null
    },
    {
      "id": 3,
      "step_order": 3,
      "role": "PurchaseOfficer",
      "status": "PENDING",
      "approver_id": 0,
      "action_at": null
    },
    {
      "id": 4,
      "step_order": 4,
      "role": "Executive",
      "status": "PENDING",
      "approver_id": 0,
      "action_at": null
    }
  ]
}
```

##### Approve Workflow Step

**Request** (As Manager approving step 2):

```http
POST http://localhost:8080/api/approval/2/approve
Authorization: Bearer <token>
Content-Type: application/json

{
  "comment": "Approved. Budget aligns with departmental goals."
}
```

**Response**: `200 OK`
```json
{
  "id": 2,
  "step_order": 2,
  "role": "Manager",
  "status": "APPROVED",
  "approver_id": 200,
  "action_at": "2026-03-08T12:15:45Z",
  "comment": "Approved. Budget aligns with departmental goals."
}
```

> **Event Published**: RabbitMQ event "approval.stepApproved" is published with workflow_id, step_order, and approver details. Purchase Service listens and updates PR status.

##### Reject Approval

```http
POST http://localhost:8080/api/approval/2/reject
Authorization: Bearer <token>
Content-Type: application/json

{
  "reason": "Budget exceeds departmental allocation for this quarter"
}
```

**Response**: `200 OK`
```json
{
  "id": 2,
  "step_order": 2,
  "role": "Manager",
  "status": "REJECTED",
  "approver_id": 200,
  "action_at": "2026-03-08T12:15:45Z",
  "reason": "Budget exceeds departmental allocation for this quarter"
}
```

> **Event Published**: RabbitMQ event "approval.rejected" is published. Purchase Service updates PR status to `REJECTED`.

##### Get Workflow by Correlation ID

```http
GET http://localhost:8080/api/approval/workflows/WF_1_20260308120530
Authorization: Bearer <token>
```

**Response**: `200 OK`
```json
{
  "id": 1,
  "entity_type": "PR",
  "entity_id": 1,
  "workflow_id": "WF_1_20260308120530",
  "status": "IN_PROGRESS",
  "current_step": 2,
  "created_by": 100,
  "created_at": "2026-03-08T12:05:31Z",
  "updated_at": "2026-03-08T12:15:45Z",
  "steps": [
    {
      "id": 1,
      "step_order": 1,
      "role": "Employee",
      "status": "APPROVED",
      "approver_id": 100,
      "action_at": "2026-03-08T12:05:31Z"
    },
    {
      "id": 2,
      "step_order": 2,
      "role": "Manager",
      "status": "APPROVED",
      "approver_id": 200,
      "action_at": "2026-03-08T12:15:45Z"
    },
    {
      "id": 3,
      "step_order": 3,
      "role": "PurchaseOfficer",
      "status": "PENDING",
      "approver_id": 0,
      "action_at": null
    },
    {
      "id": 4,
      "step_order": 4,
      "role": "Executive",
      "status": "PENDING",
      "approver_id": 0,
      "action_at": null
    }
  ]
}
```

---

## Complete Workflow Example

### Scenario: Purchase Request Approval Workflow

#### Step 1: Employee Creates PR

```http
POST http://localhost:8080/api/purchase/pr
Authorization: Bearer employee_token
Content-Type: application/json

{
  "title": "Q2 Office Equipment",
  "description": "Quarterly office supplies",
  "department": "Sales",
  "budget": 3000.00,
  "items": [
    {
      "product_name": "Monitor 27inch",
      "quantity": 5,
      "unit_price": 400.00
    }
  ]
}
```

**Result**: PR created with `status: PENDING_APPROVAL`, `workflow_id: WF_1_20260308120530`

#### Step 2: Approval Service Auto-Creates Workflow

RabbitMQ event `purchaseRequest.created` triggers approval service to create 4-step workflow.

Check approval status:

```http
GET http://localhost:8080/api/approval/PR/1
Authorization: Bearer employee_token
```

**Result**: 
- Step 1 (Employee): ✅ APPROVED (auto-approved by PR creator)
- Step 2 (Manager): ⏳ PENDING
- Step 3 (PurchaseOfficer): ⏳ PENDING
- Step 4 (Executive): ⏳ PENDING

#### Step 3: Manager Approves (Step 2)

Manager logs in and checks pending approvals:

```http
POST http://localhost:8080/api/approval/2/approve
Authorization: Bearer manager_token
Content-Type: application/json

{
  "comment": "Looks good from department side. Budget approved."
}
```

**Result**: Step 2 approved, RabbitMQ event published, workflow advances to Step 3

#### Step 4: PurchaseOfficer Approves (Step 3)

```http
POST http://localhost:8080/api/approval/3/approve
Authorization: Bearer officer_token
Content-Type: application/json

{
  "comment": "Vendor verified. Pricing competitive."
}
```

**Result**: Step 3 approved, workflow advances to Step 4

#### Step 5: Executive Approves (Step 4)

```http
POST http://localhost:8080/api/approval/4/approve
Authorization: Bearer executive_token
Content-Type: application/json

{
  "comment": "Final approval granted."
}
```

**Result**: All steps approved, workflow_status: `COMPLETED`. RabbitMQ event published.

#### Step 6: Check Final PR Status

```http
GET http://localhost:8080/api/purchase/pr/1
Authorization: Bearer employee_token
```

**Result**: PR `status: APPROVED` (updated by purchase-service listening to approval events)

#### Step 7: Create PO from Approved PR

```http
POST http://localhost:8080/api/purchase/po
Authorization: Bearer officer_token
Content-Type: application/json

{
  "pr_id": 1,
  "vendor_id": 5,
  "po_number": "PO-2026-0001",
  "expected_delivery": "2026-04-15"
}
```

**Result**: PO created, ready for fulfillment

---

## Postman Setup Guide

### 1. Import Collection

Create a new Postman Collection called **"Core Procurement API"**

### 2. Set Up Variables

In Collection → Variables, add:

```
base_url: http://localhost:8080/api
token: (leave blank, will be set after login)
```

### 3. Create Requests

#### Request 1: Login

```
Name: Auth Login
Method: POST
URL: {{base_url}}/auth/login
Body (JSON):
{
  "email": "john@company.com",
  "password": "SecurePassword123"
}

Tests Tab (Auto-capture token):
pm.environment.set("token", pm.response.json().token);
```

#### Request 2: Get Profile

```
Name: Get User Profile
Method: GET
URL: {{base_url}}/users/profile
Headers:
Key: Authorization
Value: Bearer {{token}}
```

#### Request 3: List Inventory

```
Name: List Inventory
Method: GET
URL: {{base_url}}/inventory/
Headers:
Key: Authorization
Value: Bearer {{token}}
```

#### Request 4: Create PR

```
Name: Create Purchase Request
Method: POST
URL: {{base_url}}/purchase/pr
Headers:
Authorization: Bearer {{token}}
Content-Type: application/json

Body (JSON):
{
  "title": "New PR",
  "description": "Test PR",
  "department": "Engineering",
  "budget": 5000,
  "items": [
    {
      "product_name": "Item 1",
      "quantity": 10,
      "unit_price": 100.00
    }
  ]
}

Tests Tab:
pm.environment.set("pr_id", pm.response.json().id);
pm.environment.set("workflow_id", pm.response.json().workflow_id);
```

#### Request 5: Get Approval Status

```
Name: Get Approval Status
Method: GET
URL: {{base_url}}/approval/PR/{{pr_id}}
Headers:
Authorization: Bearer {{token}}
```

#### Request 6: Approve Step

```
Name: Approve Workflow Step
Method: POST
URL: {{base_url}}/approval/2/approve
Headers:
Authorization: Bearer {{token}}
Content-Type: application/json

Body (JSON):
{
  "comment": "Approved"
}
```

### 4. Create Test Scenarios

#### Scenario: Full Workflow

1. Login (as Employee) → Get token
2. Create PR → Get pr_id and workflow_id
3. Get Approval (View 4-step workflow)
4. Login (as Manager) → Get token
5. Approve Step 2
6. Login (as PurchaseOfficer) → Get token
7. Approve Step 3
8. Login (as Executive) → Get token
9. Approve Step 4
10. Get PR (verify status changed to APPROVED)

---

## Error Responses

### 401 Unauthorized

```json
{
  "error": "unauthorized",
  "message": "missing or invalid token"
}
```

**Fix**: Include valid JWT token in Authorization header

### 403 Forbidden

```json
{
  "error": "forbidden",
  "message": "insufficient permissions for this operation"
}
```

**Fix**: Ensure user has required role (check JWT token claims)

### 404 Not Found

```json
{
  "error": "not_found",
  "message": "resource not found"
}
```

**Fix**: Verify entity ID exists and path is correct

### 400 Bad Request

```json
{
  "error": "bad_request",
  "message": "invalid fields: budget must be positive"
}
```

**Fix**: Check request body format and values

---

## Troubleshooting

### Gateway Not Responding

Check if gateway is running:

```powershell
docker ps | grep api-gateway
```

If not running, start it:

```powershell
docker compose up -d --build api-gateway
```

### "Service Unavailable" Error

The backend service may be down. Check:

```powershell
curl http://localhost:8080/api/inventory/health
curl http://localhost:8080/api/purchase/health
curl http://localhost:8080/api/approval/health
```

### Wrong Path Redirects to 404

Ensure path includes `/api/` prefix:

```
✅ Correct:   GET http://localhost:8080/api/purchase/pr/1
❌ Wrong:     GET http://localhost:8080/purchase/pr/1
```

### Token Expired

When you receive a 401 error:

1. Call login endpoint again
2. Copy new token
3. Update Authorization header in Postman with new token

---

## Quick Reference

| Service | Direct Port | Gateway Path | Prefix Stripped |
|---------|------------|--------------|-----------------|
| Auth | 6767 | `/api/auth/` or `/api/users/` | ❌ No |
| Inventory | 6768 | `/api/inventory/` or `/api/dep/` | ❌ No |
| Purchase | 6769 | `/api/purchase/` | ✅ Yes (/purchase) |
| Approval | 6770 | `/api/approval/` | ✅ Yes (/approval) |

---

## Support

For issues or questions:
1. Check service logs: `docker logs <service-name>`
2. Verify all containers running: `docker ps`
3. Review service-specific documentation in each service's API_DOCUMENTATION.md

