import { request } from "./client";
import type { User } from "./auth";

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

export const userApi = {
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
