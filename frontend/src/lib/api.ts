const BASE = "http://localhost:8080/api";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
    const res = await fetch(`${BASE}${path}`, {
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        ...options,
    });
    if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || `Request failed (${res.status})`);
    }
    return res.json();
}

export const api = {
    login: (data: { username: string; password: string }) =>
        request<{ message: string; token: string; user: User }>("/auth/login", {
            method: "POST",
            body: JSON.stringify(data),
        }),

    register: (data: {
        username: string;
        password: string;
        first_name: string;
        last_name: string;
        email: string;
    }) =>
        request<{ message: string; user: User }>("/auth/register", {
            method: "POST",
            body: JSON.stringify(data),
        }),

    logout: () =>
        request<{ message: string }>("/auth/logout", { method: "POST" }),

    me: () => request<User>("/auth/me"),

    getInventory: () => request<InventoryItem[]>("/dep/inventory"),

    getPublicInventory: () => request<PublicInventoryItem[]>("/inventory"),

    createInventory: (data: {
        sku: string;
        name: string;
        description: string;
        quantity: number;
    }) =>
        request<{ inventory: InventoryItem }>("/dep/inventory/create", {
            method: "POST",
            body: JSON.stringify(data),
        }),

    updateInventory: (id: number, data: Partial<{ sku: string; name: string; description: string; quantity: number }>) =>
        request<InventoryItem>(`/dep/inventory/${id}`, {
            method: "PATCH",
            body: JSON.stringify(data),
        }),

    deleteInventory: (id: number) =>
        request<{ message: string }>(`/dep/inventory/${id}`, { method: "DELETE" }),

    getUsers: () => request<User[]>("/users"),

    getUser: (id: number) => request<User>(`/users/${id}`),

    deleteUser: (id: number) =>
        request<{ message: string }>(`/users/${id}`, { method: "DELETE" }),

    updateUserRole: (id: number, role: string) =>
        request<{ message: string; user: User }>(`/users/${id}/role`, {
            method: "PATCH",
            body: JSON.stringify({ role }),
        }),

    updateUser: (id: number, data: { first_name: string; last_name: string; email: string }) =>
        request<{ message: string; user: User }>(`/users/${id}`, {
            method: "PUT",
            body: JSON.stringify(data),
        }),
};

export interface User {
    id: number;
    username: string;
    role: string;
    first_name: string;
    last_name: string;
    email: string;
}

export interface InventoryItem {
    ID: number;
    CreatedAt: string;
    UpdatedAt: string;
    DeletedAt: string | null;
    Sku: string;
    Name: string;
    Description: string;
    Quantity: number;
}

export interface PublicInventoryItem {
    id: number;
    sku: string;
    name: string;
    description: string;
}
