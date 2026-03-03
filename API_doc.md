# Core Procurement Service — API Documentation

## Base URLs

| Service       | Port   | Base URL                |
| ------------- | ------ | ----------------------- |
| Auth Identity | `6767` | `http://localhost:6767` |
| Inventory     | `6768` | `http://localhost:6768` |

## Roles

| Role              | Description                         |
| ----------------- | ----------------------------------- |
| `Admin`           | Full access, can manage users/roles |
| `PurchaseOfficer` | Can manage inventory (CRUD)         |
| `Manager`         | Standard authenticated user         |
| `Executive`       | Standard authenticated user         |
| `Employee`        | Default role on registration        |

## Authentication

All protected endpoints require a JWT token via **cookie** or **Authorization header**:

```
Authorization: Bearer <token>
```

---

# Auth Identity Service `:6767`

## Health Check

### `GET /health`

**Response** `200 OK`

```json
{
  "status": "ok",
  "service": "auth-identity-service"
}
```

---

## Auth

### `POST /auth/register`

Register a new user. Role defaults to `Employee`.

**Request Body**

```json
{
  "username": "john_doe",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com"
}
```

**Response** `201 Created`

```json
{
  "message": "user registered successfully",
  "user": {
    "id": 1,
    "username": "john_doe",
    "role": "Employee",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com"
  }
}
```

**Error** `409 Conflict`

```json
{
  "error": "username or email already exists"
}
```

---

### `POST /auth/login`

**Request Body**

```json
{
  "username": "john_doe",
  "password": "securePassword123"
}
```

**Response** `200 OK`

> Also sets `token` cookie (HttpOnly, 24h expiry)

```json
{
  "message": "login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "john_doe",
    "role": "Employee",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com"
  }
}
```

**Error** `401 Unauthorized`

```json
{
  "error": "invalid username or password"
}
```

---

### `POST /auth/logout`

**Response** `200 OK`

> Clears the `token` cookie.

```json
{
  "message": "logged out successfully"
}
```

---

### `GET /auth/me`

🔒 **Auth Required**

**Response** `200 OK`

```json
{
  "user_id": 1,
  "username": "john_doe",
  "role": "Employee"
}
```

---

## User Management

> All endpoints require **Auth Required** (JWT).
> Endpoints marked 🔒 **Admin** additionally require `Admin` role.

### `GET /users` 🔒 Admin

Get all users.

**Response** `200 OK`

```json
[
  {
    "id": 1,
    "username": "john_doe",
    "role": "Employee",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com"
  },
  {
    "id": 2,
    "username": "jane_smith",
    "role": "Admin",
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane.smith@example.com"
  }
]
```

---

### `GET /users/:id` 🔒 Admin

Get a specific user by ID.

**Response** `200 OK`

```json
{
  "id": 1,
  "username": "john_doe",
  "role": "Employee",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com"
}
```

**Error** `404 Not Found`

```json
{
  "error": "user not found"
}
```

---

### `PUT /users/:id`

🔒 **Auth Required**

Update user profile (any authenticated user can update their own).

**Request Body**

```json
{
  "first_name": "Johnny",
  "last_name": "Doe",
  "email": "johnny.doe@example.com"
}
```

**Response** `200 OK`

```json
{
  "message": "user updated successfully",
  "user": {
    "id": 1,
    "username": "john_doe",
    "role": "Employee",
    "first_name": "Johnny",
    "last_name": "Doe",
    "email": "johnny.doe@example.com"
  }
}
```

---

### `DELETE /users/:id` 🔒 Admin

Delete a user.

**Response** `200 OK`

```json
{
  "message": "user deleted successfully"
}
```

---

### `PATCH /users/:id/role` 🔒 Admin

Update a user's role.

**Request Body**

```json
{
  "role": "PurchaseOfficer"
}
```

**Valid roles:** `Admin`, `PurchaseOfficer`, `Employee`, `Manager`, `Executive`

**Response** `200 OK`

```json
{
  "message": "user role updated successfully",
  "user": {
    "id": 1,
    "username": "john_doe",
    "role": "PurchaseOfficer",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com"
  }
}
```

**Error** `400 Bad Request`

```json
{
  "error": "invalid role"
}
```

---

# Inventory Service `:6768`

## Health Check

### `GET /health`

**Response** `200 OK`

```json
{
  "status": "ok",
  "service": "inventory-service"
}
```

---

## Public Inventory (Authenticated Users)

### `GET /inventory`

🔒 **Auth Required** (any authenticated user)

Returns inventory list (limited view — no quantity).

**Response** `200 OK`

```json
[
  {
    "id": 1,
    "name": "A4 Paper",
    "description": "80gsm white A4 paper, 500 sheets",
    "unitprice": 150.0
  },
  {
    "id": 2,
    "name": "Ballpoint Pen",
    "description": "Blue ink ballpoint pen",
    "unitprice": 12.5
  }
]
```

---

## Department Inventory Management

> All endpoints require **PurchaseOfficer** role.

### `POST /dep/inventory/create` 🔒 PurchaseOfficer

Create a new inventory item.

**Request Body**

```json
{
  "name": "A4 Paper",
  "description": "80gsm white A4 paper, 500 sheets",
  "unitprice": 150.0,
  "quantity": 100
}
```

**Response** `201 Created`

```json
{
  "name": "A4 Paper",
  "description": "80gsm white A4 paper, 500 sheets",
  "unitprice": 150.0,
  "quantity": 100
}
```

---

### `GET /dep/inventory` 🔒 PurchaseOfficer

Get all inventory items (full details including quantity).

**Response** `200 OK`

```json
[
  {
    "ID": 1,
    "CreatedAt": "2026-03-03T10:00:00Z",
    "UpdatedAt": "2026-03-03T10:00:00Z",
    "DeletedAt": null,
    "Name": "A4 Paper",
    "Description": "80gsm white A4 paper, 500 sheets",
    "Quantity": 100,
    "UnitPrice": 150.0
  }
]
```

---

### `GET /dep/inventory/:id` 🔒 PurchaseOfficer

Get a specific inventory item.

**Response** `200 OK`

```json
{
  "ID": 1,
  "CreatedAt": "2026-03-03T10:00:00Z",
  "UpdatedAt": "2026-03-03T10:00:00Z",
  "DeletedAt": null,
  "Name": "A4 Paper",
  "Description": "80gsm white A4 paper, 500 sheets",
  "Quantity": 100,
  "UnitPrice": 150.0
}
```

**Error** `404 Not Found`

```json
{
  "error": "Inventory not found"
}
```

---

### `PATCH /dep/inventory/:id` 🔒 PurchaseOfficer

Update an inventory item.

**Request Body** (partial update)

```json
{
  "name": "A4 Paper (Updated)",
  "description": "70gsm white A4 paper, 500 sheets",
  "unitprice": 120.0,
  "quantity": 200
}
```

**Response** `200 OK`

```json
{
  "ID": 1,
  "CreatedAt": "2026-03-03T10:00:00Z",
  "UpdatedAt": "2026-03-03T12:00:00Z",
  "DeletedAt": null,
  "Name": "A4 Paper (Updated)",
  "Description": "70gsm white A4 paper, 500 sheets",
  "Quantity": 200,
  "UnitPrice": 120.0
}
```

---

### `DELETE /dep/inventory/:id` 🔒 PurchaseOfficer

Delete an inventory item (soft delete).

**Response** `200 OK`

```json
{
  "message": "Inventory deleted"
}
```

---

## Common Error Responses

| Status | Response                                                | Description                   |
| ------ | ------------------------------------------------------- | ----------------------------- |
| `400`  | `{"error": "...validation message..."}`                 | Invalid request body          |
| `401`  | `{"error": "unauthorized"}`                             | Missing token                 |
| `401`  | `{"error": "invalid or expired token"}`                 | Bad or expired JWT            |
| `403`  | `{"error": "access denied"}`                            | Insufficient role permissions |
| `404`  | `{"error": "user not found"}` / `"Inventory not found"` | Resource does not exist       |
| `409`  | `{"error": "username or email already exists"}`         | Duplicate registration        |
