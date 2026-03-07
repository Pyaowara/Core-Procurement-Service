# Purchase Service — API Documentation

## Base URLs

| Service       | Port   | Base URL                |
| ------------- | ------ | ----------------------- |
| Purchase      | `6769` | `http://localhost:6769` |

## Roles

| Role              | Description                                    |
| ----------------- | ---------------------------------------------- |
| `Admin`           | Full access, manage vendors & view all records |
| `PurchaseOfficer` | Generate POs, update PO status                 |
| `Manager`         | Create/submit PRs, receive goods               |
| `Employee`        | Create/submit own PRs                          |
| `Executive`       | (Reserved for approval workflow)               |

## Authentication

All protected endpoints require a JWT token via **cookie** or **Authorization header**:

```
Authorization: Bearer <token>
```

---

# Purchase Service `:6769`

## Health Check

### `GET /health`

**Response** `200 OK`

```json
{
  "status": "ok",
  "service": "purchase-service"
}
```

---

## Purchase Request (PR) — Management

### `POST /pr`

🔒 **Auth Required**: `Employee`, `Manager`, `Admin`

Create a new Purchase Request in DRAFT status.

**Request Body**

```json
{
  "pr_number": "000001",
  "department": "Sales",
  "items": [
    {
      "item_name": "Printer Paper",
      "description": "A4 paper 80gsm",
      "quantity": 100,
      "unit": "ชิ้น",
      "price_per_unit": 5.50,
      "discount": 10,
      "discount_unit": "%",
      "required_date": "2026-03-15"
    }
  ]
}
```

**Response** `201 Created`

```json
{
  "message": "PR created successfully",
  "data": {
    "id": 1,
    "pr_number": "000001",
    "requester_id": 1,
    "department": "Sales",
    "status": "DRAFT",
    "items": [
      {
        "id": 1,
        "item_name": "Printer Paper",
        "description": "A4 paper 80gsm",
        "quantity": 100,
        "unit": "ชิ้น",
        "price_per_unit": 5.50,
        "discount": 10,
        "discount_unit": "%",
        "total_price": 495.00,
        "required_date": "2026-03-15T00:00:00Z",
        "current_stock_at_submit": 0,
        "stock_check_at": null
      }
    ],
    "created_at": "2026-03-05T15:04:05Z",
    "updated_at": "2026-03-05T15:04:05Z"
  }
}
```

**Error** `400 Bad Request`

```json
{
  "error": "validation error message"
}
```

---

### `GET /pr`

🔒 **Auth Required**: Any authenticated user

Get all PRs for the current user (requester_id).

**Query Parameters** (optional)

| Parameter | Type   | Description              |
| --------- | ------ | ------------------------ |
| `status`  | string | Filter by status (DRAFT, PENDING, APPROVED, REJECTED) |

**Example**: `GET /pr?status=PENDING`

**Response** `200 OK`

```json
[
  {
    "id": 1,
    "pr_number": "000001",
    "requester_id": 1,
    "department": "Sales",
    "status": "PENDING",
    "workflow_id": "WF_1_20260305150405",
    "items": [...],
    "created_at": "2026-03-05T15:04:05Z"
  }
]
```

---

### `GET /pr/:id`

🔒 **Auth Required**: Any authenticated user

Get PR details by ID.

**Response** `200 OK`

```json
{
  "id": 1,
  "pr_number": "000001",
  "requester_id": 1,
  "department": "Sales",
  "status": "PENDING",
  "workflow_id": "WF_1_20260305150405",
  "items": [
    {
      "id": 1,
      "item_name": "Printer Paper",
      "quantity": 100,
      "unit": "ชิ้น",
      "price_per_unit": 5.50,
      "current_stock_at_submit": 45,
      "stock_check_at": "2026-03-05T15:04:05Z",
      "required_date": "2026-03-15T00:00:00Z"
    }
  ],
  "created_at": "2026-03-05T15:04:05Z"
}
```

**Error** `404 Not Found`

```json
{
  "error": "PR not found"
}
```

---

### `PUT /pr/:id`

🔒 **Auth Required**: `Employee`, `Manager`, `Admin`

Update PR (only DRAFT status allowed).

**Request Body** (partial update)

```json
{
  "department": "Marketing",
  "items": [
    {
      "item_name": "Printer Paper",
      "description": "A4 paper 80gsm",
      "quantity": 50,
      "unit": "ชิ้น",
      "price_per_unit": 5.50,
      "discount": 10,
      "discount_unit": "%",
      "required_date": "2026-03-20"
    }
  ]
}
```

**Response** `200 OK`

```json
{
  "message": "PR updated successfully",
  "data": {
    "id": 1,
    "pr_number": "000001",
    "department": "Marketing",
    "status": "DRAFT",
    "items": [...]
  }
}
```

**Error** `400 Bad Request`

```json
{
  "error": "only DRAFT PRs can be updated"
}
```

---

### `POST /pr/:id/submit`

🔒 **Auth Required**: `Employee`, `Manager`, `Admin`

**Submit PR for Approval** — Executes 5-step workflow:
1. **Validate Data** — Check items, required dates, quantities
2. **Check Inventory Availability** — Query Inventory Service for stock levels
3. **Take Inventory Snapshot** — Capture stock state and create snapshot
4. **Change Status** — Update PR status DRAFT → PENDING, generate WorkflowID
5. **Trigger Approval** — Publish `pr.ready.for.approval` event to Approval Service

**Request Body**: Empty (no payload required)

**Response** `200 OK`

```json
{
  "message": "PR submitted successfully",
  "data": {
    "id": 1,
    "pr_number": "000001",
    "status": "PENDING",
    "workflow_id": "WF_1_20260305150405",
    "items": [
      {
        "item_name": "Printer Paper",
        "quantity": 100,
        "unit": "ชิ้น",
        "current_stock_at_submit": 45,
        "stock_check_at": "2026-03-05T15:04:05Z",
        "required_date": "2026-03-15T00:00:00Z"
      }
    ]
  },
  "inventory_check_summary": {
    "Printer Paper": {
      "available_qty": 45,
      "checked_at": "now"
    }
  },
  "has_inventory_warnings": false,
  "snapshot_created": true,
  "workflow_id": "WF_1_20260305150405",
  "approval_event_published": true
}
```

**Error Responses**

| Status | Error Message | Description |
| ------ | ------------- | ----------- |
| `400` | `"only DRAFT PRs can be submitted"` | PR is not in DRAFT status |
| `400` | `"PR must have at least 1 item"` | PR has no items |
| `400` | `"validation failed"` | Items missing required_date or have invalid quantity |
| `404` | `"PR not found"` | PR ID does not exist |
| `500` | `"failed to create snapshot"` | Snapshot creation failed |
| `500` | `"failed to publish approval event"` | Event publishing failed |

**Example Error Response** (`400 Bad Request`)

```json
{
  "error": "validation failed",
  "details": [
    "Item 1 (Printer Paper): required_date must be specified",
    "Item 2 (Pen): quantity must be greater than 0"
  ]
}
```

---

### `GET /pr/:id/snapshot`

🔒 **Auth Required**: Any authenticated user

Get the inventory snapshot created when PR was submitted. Shows comparison between snapshot and current PR items (used for audit trail).

**Response** `200 OK`

```json
{
  "pr_id": 1,
  "pr_number": "000001",
  "status": "PENDING",
  "snapshot_created": "2026-03-05T15:04:05Z",
  "items_comparison": [
    {
      "snapshot_data": {
        "item_name": "Printer Paper",
        "quantity": 100,
        "unit": "ชิ้น",
        "price_per_unit": 5.50,
        "current_stock_at_submit": 45
      },
      "current_data": {
        "item_name": "Printer Paper",
        "quantity": 100,
        "unit": "ชิ้น",
        "price_per_unit": 5.50,
        "current_stock_at_submit": 45
      },
      "has_changed": false,
      "changed_fields": []
    }
  ],
  "summary": {
    "total_items": 1,
    "items_changed": 0
  }
}
```

**Error** `404 Not Found`

```json
{
  "error": "snapshot not found for this PR"
}
```

---

### `DELETE /pr/:id`

🔒 **Auth Required**: `Employee`, `Manager`, `Admin`

Soft delete PR (soft delete flag is set, record remains in DB for audit trail).

**Response** `200 OK`

```json
{
  "message": "PR deleted successfully"
}
```

---

## Purchase Order (PO) — Management

### `POST /po`

🔒 **Auth Required**: `PurchaseOfficer`, `Admin`

Generate a Purchase Order from an approved PR.

**Request Body**

```json
{
  "pr_id": 1,
  "vendor_id": 1,
  "credit_day": 30,
  "due_date": "2026-04-15",
  "po_items": [
    {
      "item_name": "Printer Paper",
      "description": "A4 paper 80gsm",
      "quantity": 100,
      "unit": "ชิ้น",
      "price_per_unit": 5.50,
      "discount": 10,
      "discount_unit": "%",
      "required_date": "2026-03-15"
    }
  ]
}
```

**Response** `201 Created`

```json
{
  "message": "PO created successfully",
  "data": {
    "id": 1,
    "po_number": "PO_20260305150405",
    "pr_id": 1,
    "vendor_id": 1,
    "status": "DRAFT",
    "credit_day": 30,
    "due_date": "2026-04-15T00:00:00Z",
    "items": [
      {
        "id": 1,
        "item_name": "Printer Paper",
        "description": "A4 paper 80gsm",
        "quantity": 100,
        "unit": "ชิ้น",
        "price_per_unit": 5.50,
        "discount": 10,
        "discount_unit": "%",
        "total_price": 495.00,
        "required_date": "2026-03-15T00:00:00Z"
      }
    ],
    "vendor_snapshot": {
      "vendor_id": 1,
      "vendor_name": "ABC Corporation",
      "vendor_address": "123 Business Street",
      "vendor_tax_id": "1234567890"
    },
    "created_at": "2026-03-05T15:04:05Z"
  }
}
```

**Error Responses**

| Status | Error Message | Description |
| ------ | ------------- | ----------- |
| `400` | `"only approved PRs can generate PO"` | PR is not in APPROVED status |
| `404` | `"PR not found"` | PR ID does not exist |
| `404` | `"vendor not found"` | Vendor ID does not exist |
| `500` | `"failed to create PO"` | PO creation failed |

---

### `GET /po`

🔒 **Auth Required**: Any authenticated user

Get all POs (with optional filtering).

**Query Parameters** (optional)

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `status`  | string | Filter by status (DRAFT, SENT, COMPLETED) |
| `pr_id`   | number | Filter by PR ID |

**Example**: `GET /po?status=SENT&pr_id=1`

**Response** `200 OK`

```json
[
  {
    "id": 1,
    "po_number": "PO_20260305150405",
    "pr_id": 1,
    "vendor_id": 1,
    "status": "DRAFT",
    "credit_day": 30,
    "due_date": "2026-04-15T00:00:00Z",
    "items": [...]
  }
]
```

---

### `GET /po/:id`

🔒 **Auth Required**: Any authenticated user

Get PO details by ID (includes vendor snapshot).

**Response** `200 OK`

```json
{
  "id": 1,
  "po_number": "PO_20260305150405",
  "pr_id": 1,
  "vendor_id": 1,
  "status": "DRAFT",
  "credit_day": 30,
  "due_date": "2026-04-15T00:00:00Z",
  "items": [
    {
      "id": 1,
      "item_name": "Printer Paper",
      "quantity": 100,
      "unit": "ชิ้น",
      "price_per_unit": 5.50,
      "total_price": 495.00
    }
  ],
  "vendor_snapshot": {
    "vendor_id": 1,
    "vendor_name": "ABC Corporation",
    "vendor_address": "123 Business Street",
    "vendor_tax_id": "1234567890"
  },
  "created_at": "2026-03-05T15:04:05Z"
}
```

**Error** `404 Not Found`

```json
{
  "error": "PO not found"
}
```

---

### `PUT /po/:id`

🔒 **Auth Required**: `PurchaseOfficer`, `Admin`

Update PO status or other details.

**Request Body** (partial update)

```json
{
  "status": "SENT",
  "credit_day": 45,
  "due_date": "2026-04-20"
}
```

**Valid Status Values**: `DRAFT`, `SENT`, `COMPLETED`

**Response** `200 OK`

```json
{
  "message": "PO updated successfully",
  "data": {
    "id": 1,
    "po_number": "PO_20260305150405",
    "status": "SENT",
    "credit_day": 45,
    "due_date": "2026-04-20T00:00:00Z"
  }
}
```

**Error** `400 Bad Request`

```json
{
  "error": "invalid status"
}
```

---

### `POST /po/:id/receive`

🔒 **Auth Required**: `Manager`, `PurchaseOfficer`, `Admin`

Record goods reception and publish event to Inventory Service for stock update.

**Request Body**

```json
{
  "received_qty": {
    "Printer Paper": 100
  }
}
```

**Response** `200 OK`

```json
{
  "message": "Goods received successfully",
  "data": {
    "id": 1,
    "po_id": 1,
    "received_data": "{\"Printer Paper\": 100}",
    "received_at": "2026-03-05T15:04:05Z"
  }
}
```

**Note**: Event `goods.received` is published to Inventory Service to update stock.

---

### `DELETE /po/:id`

🔒 **Auth Required**: `PurchaseOfficer`, `Admin`

Soft delete PO.

**Response** `200 OK`

```json
{
  "message": "PO deleted successfully"
}
```

---

## Vendor Management

### `POST /vendor`

🔒 **Auth Required**: `Admin` only

Create a new vendor.

**Request Body**

```json
{
  "vendor_name": "ABC Corporation",
  "vendor_address": "123 Business Street, Bangkok",
  "vendor_tax_id": "1234567890"
}
```

**Response** `201 Created`

```json
{
  "message": "Vendor created successfully",
  "data": {
    "id": 1,
    "vendor_name": "ABC Corporation",
    "vendor_address": "123 Business Street, Bangkok",
    "vendor_tax_id": "1234567890",
    "created_at": "2026-03-05T15:04:05Z"
  }
}
```

**Error** `400 Bad Request`

```json
{
  "error": "vendor_tax_id must be unique"
}
```

---

### `GET /vendor`

🔒 **Auth Required**: Any authenticated user

Get all vendors.

**Response** `200 OK`

```json
[
  {
    "id": 1,
    "vendor_name": "ABC Corporation",
    "vendor_address": "123 Business Street, Bangkok",
    "vendor_tax_id": "1234567890",
    "created_at": "2026-03-05T15:04:05Z"
  }
]
```

---

### `GET /vendor/:id`

🔒 **Auth Required**: Any authenticated user

Get vendor details by ID.

**Response** `200 OK`

```json
{
  "id": 1,
  "vendor_name": "ABC Corporation",
  "vendor_address": "123 Business Street, Bangkok",
  "vendor_tax_id": "1234567890",
  "created_at": "2026-03-05T15:04:05Z"
}
```

**Error** `404 Not Found`

```json
{
  "error": "vendor not found"
}
```

---

### `PUT /vendor/:id`

🔒 **Auth Required**: `Admin` only

Update vendor information.

**Request Body** (partial update)

```json
{
  "vendor_name": "ABC Corporation (Updated)",
  "vendor_address": "456 New Street, Bangkok"
}
```

**Response** `200 OK`

```json
{
  "message": "Vendor updated successfully",
  "data": {
    "id": 1,
    "vendor_name": "ABC Corporation (Updated)",
    "vendor_address": "456 New Street, Bangkok",
    "vendor_tax_id": "1234567890"
  }
}
```

---

### `DELETE /vendor/:id`

🔒 **Auth Required**: `Admin` only

Delete vendor.

**Response** `200 OK`

```json
{
  "message": "Vendor deleted successfully"
}
```

---

## Message Events

### Published Events

#### 1. `pr.ready.for.approval` (Topic)

Published when PR is submitted for approval via `POST /pr/:id/submit`.

**Payload**

```json
{
  "pr_id": 1,
  "pr_number": "000001",
  "requester_id": 1,
  "department": "Sales",
  "items": [
    {
      "item_name": "Printer Paper",
      "description": "A4 paper 80gsm",
      "quantity": 100,
      "unit": "ชิ้น",
      "price_per_unit": 5.50,
      "discount": 10,
      "discount_unit": "%",
      "total_price": 495.00,
      "required_date": "2026-03-15"
    }
  ],
  "workflow_id": "WF_1_20260305150405",
  "timestamp": "2026-03-05T15:04:05Z"
}
```

**Consumer**: Approval Service → Creates approval workflow

---

#### 2. `po.created` (Topic)

Published when PO is created via `POST /po`.

**Payload**

```json
{
  "po_id": 1,
  "po_number": "PO_20260305150405",
  "pr_id": 1,
  "vendor_id": 1,
  "vendor_name": "ABC Corporation",
  "items": [...],
  "due_date": "2026-04-15",
  "timestamp": "2026-03-05T15:04:05Z"
}
```

---

#### 3. `goods.received` (Topic)

Published when goods are received via `POST /po/:id/receive`.

**Payload**

```json
{
  "po_id": 1,
  "po_number": "PO_20260305150405",
  "received_qty": {
    "Printer Paper": 100
  },
  "timestamp": "2026-03-05T15:04:05Z"
}
```

**Consumer**: Inventory Service → Updates stock

---

### Subscribed Events

#### 1. `approval.completed` (From Approval Service)

Updates PR status to APPROVED when approval is completed.

**Payload**

```json
{
  "pr_id": 1,
  "workflow_id": "WF_1_20260305150405",
  "status": "approved",
  "approved_at": "2026-03-05T15:04:05Z"
}
```

---

#### 2. `approval.rejected` (From Approval Service)

Updates PR status to REJECTED when approval is rejected.

**Payload**

```json
{
  "pr_id": 1,
  "workflow_id": "WF_1_20260305150405",
  "reason": "Budget exceeded",
  "rejected_at": "2026-03-05T15:04:05Z"
}
```

---

## Common Error Responses

| Status | Response | Description |
| ------ | -------- | ----------- |
| `400` | `{"error": "..."}` | Invalid request body or validation failed |
| `401` | `{"error": "unauthorized"}` | Missing token |
| `401` | `{"error": "invalid or expired token"}` | Bad or expired JWT |
| `403` | `{"error": "access denied"}` | Insufficient role permissions |
| `404` | `{"error": "...not found"}` | Resource does not exist |
| `500` | `{"error": "..."}` | Server error |

---

## Data Models

### Purchase Request (PR)

| Field | Type | Description |
| ----- | ---- | ----------- |
| `id` | uint | Primary key |
| `pr_number` | string | Unique PR number |
| `requester_id` | uint | User who requested |
| `department` | string | Department name |
| `status` | string | DRAFT, PENDING, APPROVED, REJECTED |
| `workflow_id` | string | Link to Approval Service workflow |
| `is_deleted` | bool | Soft delete flag |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |

### PR Item

| Field | Type | Description |
| ----- | ---- | ----------- |
| `id` | uint | Primary key |
| `pr_id` | uint | Foreign key to PR |
| `item_name` | string | Item name (snapshot at submission) |
| `description` | string | Item description |
| `quantity` | int | Quantity requested |
| `unit` | string | Unit of measurement (ชิ้น, แท่ง, etc.) |
| `price_per_unit` | decimal | Price per unit (snapshot at submission) |
| `discount` | decimal | Discount amount |
| `discount_unit` | string | % or BAHT |
| `total_price` | decimal | Total price |
| `required_date` | timestamp | Date item is needed |
| `current_stock_at_submit` | int | **NEW** Stock available at submission time |
| `stock_check_at` | timestamp | **NEW** When stock was checked |

### Purchase Order (PO)

| Field | Type | Description |
| ----- | ---- | ----------- |
| `id` | uint | Primary key |
| `po_number` | string | Unique PO number |
| `pr_id` | uint | Reference to PR |
| `vendor_id` | uint | Reference to vendor |
| `status` | string | DRAFT, SENT, COMPLETED |
| `credit_day` | int | Payment terms (days) |
| `due_date` | timestamp | Expected delivery date |
| `is_deleted` | bool | Soft delete flag |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |

### PO Item

| Field | Type | Description |
| ----- | ---- | ----------- |
| `id` | uint | Primary key |
| `po_id` | uint | Foreign key to PO |
| `item_name` | string | Item name |
| `description` | string | Item description |
| `quantity` | int | Quantity ordered |
| `unit` | string | Unit of measurement |
| `price_per_unit` | decimal | Price per unit |
| `discount` | decimal | Discount amount |
| `discount_unit` | string | % or BAHT |
| `total_price` | decimal | Total price |
| `required_date` | timestamp | Date item is needed |

### Vendor

| Field | Type | Description |
| ----- | ---- | ----------- |
| `id` | uint | Primary key |
| `vendor_name` | string | Vendor company name |
| `vendor_address` | string | Vendor address |
| `vendor_tax_id` | string | Tax ID (unique) |

### Data Snapshots

#### Inventory Snapshot (InventorySnapshot)

| Field | Type | Description |
| ----- | ---- | ----------- |
| `id` | uint | Primary key |
| `pr_id` | uint | Foreign key to PR (unique) |
| `snapshot_data` | JSON | JSON of PR Items at submission time |
| `created_at` | timestamp | When snapshot was created |

**Purpose**: Audit trail, data consistency, change detection

#### Vendor Snapshot (VendorSnapshot)

| Field | Type | Description |
| ----- | ---- | ----------- |
| `id` | uint | Primary key |
| `po_id` | uint | Foreign key to PO |
| `vendor_id` | uint | Vendor ID |
| `vendor_name` | string | Vendor name (snapshot at PO creation) |
| `vendor_address` | string | Vendor address (snapshot) |
| `vendor_tax_id` | string | Vendor tax ID (snapshot) |
| `snapshot_data` | JSON | Full vendor data snapshot |

**Purpose**: Data consistency — vendor info doesn't change retroactively

---

## Example Workflow

### Complete PR → PO Procurement Flow

```
1. Employee creates PR (POST /pr)
   ├─ Status: DRAFT
   ├─ Items: List of needed items
   └─ Stored in database

2. Employee submits PR for approval (POST /pr/:id/submit)
   ├─ STEP 1: Validate data (required_date, quantity)
   ├─ STEP 2: Check inventory with Inventory Service
   ├─ STEP 3: Create snapshot with stock info
   ├─ STEP 4: PR Status: DRAFT → PENDING, generate WorkflowID
   ├─ STEP 5: Publish pr.ready.for.approval event
   └─ Response includes stock availability summary

3. Approval Service processes event (async)
   ├─ Creates approval workflow
   ├─ Manager/Executive reviews PR
   └─ Publishes approval.completed or approval.rejected

4. Purchase Service receives approval event (subscribed)
   ├─ If approval.completed: PR Status → APPROVED
   └─ If approval.rejected: PR Status → REJECTED

5. Purchase Officer generates PO (POST /po)
   ├─ Only from APPROVED PR
   ├─ Select vendor
   ├─ Create VendorSnapshot
   ├─ Publish po.created event
   └─ Status: DRAFT

6. Purchase Officer updates PO status (PUT /po/:id)
   ├─ Status: DRAFT → SENT
   └─ Vendor notified (custom logic)

7. Manager records goods reception (POST /po/:id/receive)
   ├─ Record received quantities
   ├─ Create GoodsReceived record
   ├─ Publish goods.received event
   └─ Event consumed by Inventory Service to update stock

8. Inventory Service updates stock (subscribed event)
   ├─ Receives goods.received event
   ├─ Updates inventory quantities
   └─ Records transaction
```

---

## Technology Stack

- **Language**: Go 1.25.0
- **Web Framework**: Gin v1.11.0
- **Database**: PostgreSQL with GORM v1.31.1
- **Message Broker**: RabbitMQ v3.13 (AMQP 0.9.1)
- **Authentication**: JWT v5.3.1
- **Environment**: godotenv v1.5.1

---

## Environment Variables

```env
PORT=6769
DB_DSN=host=localhost user=postgres password=postgres dbname=purchase_db port=5435 sslmode=disable
JWT_SECRET=super-secret-jwt-key-change-me
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
INVENTORY_SERVICE_URL=http://localhost:6768
```
