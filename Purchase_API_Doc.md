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

## Key Features

- **Auto-Generated PR Numbers**: Format `PR-YYYYMMDD-{PRID:06d}` (e.g., PR-20260305-000001)
- **Inventory Integration**: Stock levels checked during PR creation/update
- **Smart PR Item Creation**: Only PR items with insufficient stock are created
- **Snapshot-Based Auditing**: Captures state at PR submission and PO creation for audit trails
- **Event-Driven Architecture**: RabbitMQ integration for inter-service communication

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

Create a new Purchase Request in DRAFT status using an **atomic transaction**. PR number is auto-generated if not provided (format: `PR-YYYYMMDD-{PRID:06d}`).

**Transaction Behavior**: 
- PR creation, item validation, and PR items creation are atomic using database transaction
- If any step fails or no items are created, the entire transaction is rolled back
- No orphaned PR records will exist without items

**Key Behavior**: 
- Checks inventory stock for each requested item
- Only creates PR items when stock is **insufficient**
- PR item quantity = (requested qty - available qty)
- If item not found in inventory: creates PR for full requested qty

**Request Body**

```json
{
  "pr_number": "PR-20260305-000001",
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

**Note**: `pr_number` is optional. If omitted, it will be auto-generated.

**Response** `201 Created`

```json
{
  "message": "PR created successfully",
  "data": {
    "id": 1,
    "pr_number": "PR-20260305-000001",
    "requester_id": 1,
    "department": "Sales",
    "status": "DRAFT",
    "workflow_id": null,
    "items": [
      {
        "id": 1,
        "item_name": "Printer Paper",
        "description": "A4 paper 80gsm",
        "quantity": 50,
        "unit": "ชิ้น",
        "price_per_unit": 5.50,
        "discount": 10,
        "discount_unit": "%",
        "total_price": 242.50,
        "required_date": "2026-03-15T00:00:00Z",
        "current_stock_at_submit": 0,
        "stock_check_at": null
      }
    ],
    "created_at": "2026-03-05T15:04:05Z",
    "updated_at": "2026-03-05T15:04:05Z"
  },
  "inventory_check_summary": {
    "Printer Paper": {
      "requested_qty": 100,
      "available_qty": 50,
      "pr_qty_created": 50,
      "status": "insufficient stock (50 available), creating PR for shortage of 50 units"
    }
  },
  "pr_items_created_count": 1,
  "total_items_requested": 1,
  "items_with_sufficient_qty": 0
}
```

**Error Responses**

| Status | Error Message | Description |
| ------ | ------------- | ----------- |
| `400` | `"no items to create PR - all items have sufficient stock"` | All requested items have sufficient inventory; PR is not created |
| `400` | `"validation error message"` | Request validation failed |

**Note**: Transaction ensures that if any error occurs, the PR and all items are rolled back. Either the complete PR with items is created, or nothing at all.

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
    "pr_number": "PR-20260305-000001",
    "requester_id": 1,
    "department": "Sales",
    "status": "PENDING",
    "workflow_id": "WF_1_20260305150405",
    "items": [
      {
        "id": 1,
        "item_name": "Printer Paper",
        "quantity": 50,
        "unit": "ชิ้น",
        "price_per_unit": 5.50,
        "required_date": "2026-03-15T00:00:00Z"
      }
    ],
    "created_at": "2026-03-05T15:04:05Z",
    "updated_at": "2026-03-05T15:04:05Z"
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
  "pr_number": "PR-20260305-000001",
  "requester_id": 1,
  "department": "Sales",
  "status": "PENDING",
  "workflow_id": "WF_1_20260305150405",
  "items": [
    {
      "id": 1,
      "item_name": "Printer Paper",
      "quantity": 50,
      "unit": "ชิ้น",
      "price_per_unit": 5.50,
      "current_stock_at_submit": 45,
      "stock_check_at": "2026-03-05T15:04:05Z",
      "required_date": "2026-03-15T00:00:00Z"
    }
  ],
  "created_at": "2026-03-05T15:04:05Z",
  "updated_at": "2026-03-05T15:04:05Z"
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

Update PR (only DRAFT status allowed) using an **atomic transaction**. Works the same as CreatePR with inventory checking and smart PR item creation based on stock availability.

**Transaction Behavior**:
- Department update and item changes are atomic
- If any step fails or no items are created, the entire transaction is rolled back
- All changes succeed together or none at all

**Request Body** (partial update)

```json
{
  "department": "Marketing",
  "items": [
    {
      "item_name": "Printer Paper",
      "description": "A4 paper 80gsm",
      "quantity": 200,
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
    "pr_number": "PR-20260305-000001",
    "department": "Marketing",
    "status": "DRAFT",
    "items": [...]
  },
  "inventory_check_summary": {
    "Printer Paper": {
      "requested_qty": 200,
      "available_qty": 45,
      "pr_qty_created": 155,
      "status": "insufficient stock (45 available), creating PR for shortage of 155 units"
    }
  },
  "pr_items_created_count": 1,
  "total_items_requested": 1,
  "items_with_sufficient_qty": 0
}
```

**Error** `400 Bad Request`

```json
{
  "error": "only DRAFT PRs can be updated"
}
```

Other errors:
- `"no items to update PR - all items have sufficient stock"` — All requested items have sufficient stock; PR items are not updated
- `"PR not found"` — PR with given ID does not exist

**Note**: Transaction ensures atomicity. Department update and item changes are committed together, or rolled back together if any step fails.

---

### `POST /pr/:id/submit`

🔒 **Auth Required**: `Employee`, `Manager`, `Admin`

**Submit PR for Approval** — Executes workflow:
1. **Validate Data** — Check items and required fields
2. **Create Inventory Snapshot** — Capture PR state for audit trail
3. **Change Status** — PR status DRAFT → PENDING, generate WorkflowID
4. **Trigger Approval** — Publish `pr.ready.for.approval` event to Approval Service

**NOTE**: Inventory checking happens during CreatePR/UpdatePR, not here.

**Request Body**: Empty (no payload required)

**Response** `200 OK`

```json
{
  "message": "PR submitted successfully",
  "data": {
    "id": 1,
    "pr_number": "PR-20260305-000001",
    "status": "PENDING",
    "workflow_id": "WF_1_20260305150405",
    "items": [
      {
        "item_name": "Printer Paper",
        "quantity": 50,
        "unit": "ชิ้น",
        "required_date": "2026-03-15T00:00:00Z"
      }
    ]
  },
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

---

### `GET /pr/:id/snapshot`

🔒 **Auth Required**: Any authenticated user

Retrieves the inventory snapshot created when PR was submitted. Shows comparison between snapshot data and current PR items for audit trail.

**Response** `200 OK`

```json
{
  "pr_id": 1,
  "pr_number": "PR-20260305-000001",
  "status": "PENDING",
  "snapshot_created": "2026-03-05T15:04:05Z",
  "items_comparison": [
    {
      "snapshot_data": {
        "id": 1,
        "item_name": "Printer Paper",
        "quantity": 50,
        "unit": "ชิ้น",
        "price_per_unit": 5.50,
        "required_date": "2026-03-15T00:00:00Z"
      },
      "current_data": {
        "id": 1,
        "item_name": "Printer Paper",
        "quantity": 50,
        "unit": "ชิ้น",
        "price_per_unit": 5.50,
        "required_date": "2026-03-15T00:00:00Z"
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

Soft delete a PR. 

**Response** `200 OK`

```json
{
  "message": "PR deleted successfully"
}
```

**Error** `404 Not Found`

```json
{
  "error": "PR not found"
}
```

---

## Purchase Order (PO) — Management

### `POST /po`

🔒 **Auth Required**: `PurchaseOfficer`, `Admin`

Create a Purchase Order from an approved PR using an **atomic transaction**. PO items are automatically sourced from the selected PR items. Multiple POs can be created from the same PR.

**Transaction Behavior**: 
- PO creation, PO items creation, and vendor snapshot creation are atomic
- If any step fails, the entire transaction is rolled back
- No orphaned PO records will exist without items

**Request Body**

```json
{
  "pr_id": 1,
  "vendor_id": 1,
  "credit_day": 30,
  "due_date": "2026-04-15",
  "item_ids": [1, 2, 3]
}
```

**Parameters**

| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `pr_id` | integer | ✓ | ID of approved PR |
| `vendor_id` | integer | ✓ | ID of vendor |
| `credit_day` | integer | | Payment terms in days (default: 0) |
| `due_date` | string (YYYY-MM-DD) | ✓ | Expected delivery date |
| `item_ids` | array[integer] | | Array of PR item IDs to include. All IDs must belong to the specified PR. If omitted, all PR items will be used. |

**Validation Rules**

- All PR item IDs in `item_ids` must belong to the specified PR
- At least 1 valid PO item must be created (otherwise PO is not created)
- Multiple POs can be created from the same PR with different vendors/items

**Examples**

Create PO with all PR items:
```json
{
  "pr_id": 1,
  "vendor_id": 1,
  "credit_day": 30,
  "due_date": "2026-04-15"
}
```

Create PO with selected items:
```json
{
  "pr_id": 1,
  "vendor_id": 1,
  "credit_day": 30,
  "due_date": "2026-04-15",
  "item_ids": [1, 3]
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
        "po_id": 1,
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
    "created_at": "2026-03-05T15:04:05Z",
    "updated_at": "2026-03-05T15:04:05Z"
  }
}
```

**Error Responses**

| Status | Error Message | Description |
| ------ | ------------- | ----------- |
| `400` | `"no valid items to create PO"` | No items selected or all selected items are invalid; PO is not created |
| `400` | `"PR item ID X does not belong to this PR"` | Requested item ID not in PR |
| `400` | `"only approved PRs can generate PO"` | PR is not in APPROVED status |
| `404` | `"PR not found"` | PR ID does not exist |
| `404` | `"vendor not found"` | Vendor ID does not exist |
| `500` | `"failed to create PO and items"` | Transaction failed during creation |

**Note**: Uses **atomic transaction**. PO creation, PO items creation, and vendor snapshot creation either all succeed or all fail. No orphaned PO records will be created without items.

---

### `GET /po`

🔒 **Auth Required**: Any authenticated user

Get all POs with optional filters.

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
    "items": [...],
    "created_at": "2026-03-05T15:04:05Z"
  }
]
```

---

### `GET /po/:id`

🔒 **Auth Required**: Any authenticated user

Get PO details by ID, including vendor snapshot.

**Response** `200 OK`

```json
{
  "id": 1,
  "po_number": "PO_20260305150405",
  "pr_id": 1,
  "vendor_id": 1,
  "status": "SENT",
  "credit_day": 30,
  "due_date": "2026-04-15T00:00:00Z",
  "items": [
    {
      "id": 1,
      "po_id": 1,
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
    "id": 1,
    "po_id": 1,
    "vendor_id": 1,
    "vendor_name": "ABC Supplies",
    "vendor_address": "123 Business St",
    "vendor_tax_id": "1234567890",
    "snapshot_data": {...},
    "created_at": "2026-03-05T15:04:05Z"
  },
  "created_at": "2026-03-05T15:04:05Z",
  "updated_at": "2026-03-05T15:04:05Z"
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

Update PO status and other details.

**Request Body**

```json
{
  "status": "SENT",
  "credit_day": 45,
  "due_date": "2026-04-30"
}
```

**Valid Statuses**: `DRAFT`, `SENT`, `COMPLETED`

**Response** `200 OK`

```json
{
  "message": "PO updated successfully",
  "data": {
    "id": 1,
    "po_number": "PO_20260305150405",
    "status": "SENT",
    "credit_day": 45,
    "due_date": "2026-04-30T00:00:00Z",
    "items": [...],
    "updated_at": "2026-03-05T15:10:00Z"
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

Record goods reception for a PO. Publishes `goods.received` event to Inventory Service.

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
    "received_data": {
      "Printer Paper": 100
    },
    "received_at": "2026-03-05T15:15:00Z",
    "created_at": "2026-03-05T15:15:00Z"
  }
}
```

**Error** `404 Not Found`

```json
{
  "error": "PO not found"
}
```

---

### `DELETE /po/:id`

🔒 **Auth Required**: `PurchaseOfficer`, `Admin`

Soft delete a PO.

**Response** `200 OK`

```json
{
  "message": "PO deleted successfully"
}
```

**Error** `404 Not Found`

```json
{
  "error": "PO not found"
}
```

---

## Vendor Management

### `POST /vendor`

🔒 **Auth Required**: `Admin`

Create a new vendor.

**Request Body**

```json
{
  "name": "ABC Supplies",
  "address": "123 Business Street, Bangkok 10110",
  "tax_id": "1234567890123"
}
```

**Response** `201 Created`

```json
{
  "message": "Vendor created successfully",
  "data": {
    "id": 1,
    "name": "ABC Supplies",
    "address": "123 Business Street, Bangkok 10110",
    "tax_id": "1234567890123",
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

### `GET /vendor`

🔒 **Auth Required**: Any authenticated user

Get all vendors.

**Response** `200 OK`

```json
[
  {
    "id": 1,
    "name": "ABC Supplies",
    "address": "123 Business Street, Bangkok 10110",
    "tax_id": "1234567890123",
    "created_at": "2026-03-05T15:04:05Z",
    "updated_at": "2026-03-05T15:04:05Z"
  },
  {
    "id": 2,
    "name": "XYZ Trading",
    "address": "456 Trade Avenue, Bangkok 10120",
    "tax_id": "9876543210987",
    "created_at": "2026-03-04T10:00:00Z",
    "updated_at": "2026-03-04T10:00:00Z"
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
  "name": "ABC Supplies",
  "address": "123 Business Street, Bangkok 10110",
  "tax_id": "1234567890123",
  "created_at": "2026-03-05T15:04:05Z",
  "updated_at": "2026-03-05T15:04:05Z"
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

🔒 **Auth Required**: `Admin`

Update vendor information.

**Request Body**

```json
{
  "name": "ABC Supplies Co., Ltd.",
  "address": "123 Business Street, Bangkok 10110, Thailand",
  "tax_id": "1234567890123"
}
```

**Response** `200 OK`

```json
{
  "message": "Vendor updated successfully",
  "data": {
    "id": 1,
    "name": "ABC Supplies Co., Ltd.",
    "address": "123 Business Street, Bangkok 10110, Thailand",
    "tax_id": "1234567890123",
    "updated_at": "2026-03-05T15:20:00Z"
  }
}
```

**Error** `404 Not Found`

```json
{
  "error": "vendor not found"
}
```

---

### `DELETE /vendor/:id`

🔒 **Auth Required**: `Admin`

Delete a vendor.

**Response** `200 OK`

```json
{
  "message": "Vendor deleted successfully"
}
```

**Error** `404 Not Found`

```json
{
  "error": "vendor not found"
}
```

---

## API Endpoints Summary

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/pr` | Create new PR | Employee, Manager, Admin |
| GET | `/pr` | List user's PRs | Authenticated |
| GET | `/pr/:id` | Get PR details | Authenticated |
| PUT | `/pr/:id` | Update PR (DRAFT only) | Employee, Manager, Admin |
| POST | `/pr/:id/submit` | Submit PR for approval | Employee, Manager, Admin |
| GET | `/pr/:id/snapshot` | Get PR audit snapshot | Authenticated |
| DELETE | `/pr/:id` | Delete PR | Employee, Manager, Admin |
| POST | `/po` | Create PO from approved PR | PurchaseOfficer, Admin |
| GET | `/po` | List all POs | Authenticated |
| GET | `/po/:id` | Get PO details | Authenticated |
| PUT | `/po/:id` | Update PO status | PurchaseOfficer, Admin |
| POST | `/po/:id/receive` | Record goods reception | Manager, PurchaseOfficer, Admin |
| DELETE | `/po/:id` | Delete PO | PurchaseOfficer, Admin |
| POST | `/vendor` | Create vendor | Admin |
| GET | `/vendor` | List all vendors | Authenticated |
| GET | `/vendor/:id` | Get vendor details | Authenticated |
| PUT | `/vendor/:id` | Update vendor | Admin |
| DELETE | `/vendor/:id` | Delete vendor | Admin |

---

## Event Publishing

The Purchase Service publishes the following events to RabbitMQ:

| Event | Published When | Payload |
|-------|---|---------|
| `pr.ready.for.approval` | PR is submitted for approval | PR details, items, workflow ID |
| `po.created` | PO is created from approved PR | PO details, vendor info, items |
| `goods.received` | Goods are received for a PO | PO ID, received quantities |

---

## Approval Workflow

### PR Status Transitions

```
DRAFT → (submit) → PENDING → (approval event) → APPROVED
                         ↓                         ↓
                    (reject event)             (generate PO)
                         ↓                         ↓
                      REJECTED                 PO Status
```

### Approval Event Handling

The Purchase Service auto-subscribes to approval events:

- **approval.completed**: Changes PR status DRAFT → APPROVED
- **approval.rejected**: Changes PR status DRAFT → REJECTED

---

## Data Model Reference

### PR Statuses
- `DRAFT`: Initial state, can be edited/deleted
- `PENDING`: Submitted for approval, awaiting approval service
- `APPROVED`: Approved by approval workflow
- `REJECTED`: Rejected by approval workflow

### PO Statuses
- `DRAFT`: Initial state created from approved PR
- `SENT`: Sent to vendor
- `COMPLETED`: Goods fully received

### Inventory Checking Logic

**During CreatePR/UpdatePR**:
1. Check current inventory stock for each item via Inventory Service
2. **If available_qty ≥ requested_qty**: No PR item created (sufficient stock)
3. **If available_qty < requested_qty**: Create PR item with qty = (requested - available)
4. **If item not found**: Create PR item with full requested qty

This smart logic reduces unnecessary procurement by only creating PR items when there's actual stock shortage, ensuring efficient resource management.

---

## Example Workflow

### 1. Create PR with Inventory Checking
```
POST /pr
{
  "department": "IT",
  "items": [
    {"item_name": "Laptop", "quantity": 10, ...}
  ]
}
```
- Checks: 10 Laptops requested, 7 available
- Response: Creates PR with 3 Laptops (shortage)

### 2. Submit PR for Approval
```
POST /pr/:id/submit
```
- Creates snapshot of current PR state
- Publishes event to Approval Service
- Status changes: DRAFT → PENDING

### 3. Approval Service Approves (event-driven)
- Receives approval event
- Updates PR: PENDING → APPROVED

### 4. Generate PO from Approved PR
```
POST /po
{
  "pr_id": 1,
  "vendor_id": 1,
  "po_items": [...]
}
```
- Creates PO in DRAFT status
- Creates vendor snapshot
- Publishes PO_CREATED event

### 5. Send PO and Record Goods Reception
```
PUT /po/:id
{"status": "SENT"}

POST /po/:id/receive
{"received_qty": {"Laptop": 3}}
```
- Updates status and notifies Inventory Service
- Inventory Service updates stock levels
