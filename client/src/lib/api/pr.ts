import { request } from "./client";

export interface PRItem {
    ID: number;
    PRID: number;
    CreatedAt: string;
    UpdatedAt: string;
    SKU: string;
    ItemName: string;
    Description: string;
    Quantity: number;
    PricePerUnit: number;
    Discount: number;
    DiscountUnit: string;
    TotalPrice: number;
    RequiredDate: string; // Optional flag to indicate if the item is marked for deletion
}

export interface PurchaseRequest {
    ID: number;
    CreatedAt: string;
    UpdatedAt: string;
    DeletedAt: string | null;
    PRNumber: string;
    Purpose: string;
    Department: string;
    Status: string;
    RequesterID: number;
    WorkflowID: string;
    IsDeleted: boolean;
    Items: PRItem[];
}

export const prApi = {
    getPurchaseRequests: () => request<PurchaseRequest[]>("/purchase/pr"),

    getPurchaseRequestById: (id: number) => request<PurchaseRequest>(`/purchase/pr/${id}`),

    createPurchaseRequest: (data: {
        department: string;
        purpose: string;
        items: Array<{
            sku: string;
            item_name: string;
            description: string;
            quantity: number;
            price_per_unit: number;
            discount?: number;
            discount_unit?: string;
            required_date: string;
        }>;
    }) =>
        request<{ message: string; data: PurchaseRequest }>("/purchase/pr", {
            method: "POST",
            body: JSON.stringify(data),
        }),

    editPurchaseRequest: (id: number, data: {
        department: string;
        purpose: string;
        items: Array<{
            sku: string;
            item_name: string;
            description: string;
            quantity: number;
            price_per_unit: number;
            discount?: number;
            discount_unit?: string;
            required_date: string;
        }>;
    }) =>
        request<{ message: string; data: PurchaseRequest }>(`/purchase/pr/${id}`, {
            method: "PUT",
            body: JSON.stringify(data),
        }),

    submitPurchaseRequest: (id: number) =>
        request<{ message: string; data: PurchaseRequest }>(`/purchase/pr/${id}/submit`, {
            method: "POST",
        }),

    deletePurchaseRequest: (id: number) =>
        request<{ message: string }>(`/purchase/pr/${id}`, {
            method: "DELETE",
        }),

};
