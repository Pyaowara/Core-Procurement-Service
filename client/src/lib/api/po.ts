import { request } from "./client";

export interface POItem {
    ID: number;
    POID: number;
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
    RequiredDate: string;
}

export interface Vendor {
    ID: number;
    Name: string;
    Address: string;
    TaxID: string;
    CreatedAt: string;
    UpdatedAt: string;
}

export interface PurchaseOrder {
    ID: number;
    CreatedAt: string;
    UpdatedAt: string;
    DeletedAt: string | null;
    PONumber: string;
    PRID: number;
    VendorID: number;
    Vendor?: Vendor;
    Purpose: string;
    Status: string;
    CreditDay: number;
    DueDate: string;
    IsDeleted: boolean;
    Items: POItem[];
}

export const poApi = {
    getPurchaseOrders: () => {
        
        return request<PurchaseOrder[]>(`/purchase/po`);
    },

    getPurchaseOrderById: (id: number) => request<PurchaseOrder>(`/purchase/po/${id}`),

    createPurchaseOrder: (data: {
        pr_id: number;
        vendor_id: number;
        po_items: Array<{
            pr_item_id: number;
            sku?: string;
            item_name?: string;
            description?: string;
            quantity?: number;
            price_per_unit?: number;
            discount?: number;
            discount_unit?: string;
            required_date?: string;
        }>;
        credit_day?: number;
        due_date: string;
    }) =>
        request<{ message: string; data: PurchaseOrder }>("/purchase/po", {
            method: "POST",
            body: JSON.stringify(data),
        }),

    receiveGoods: (id: number) =>
        request<{ message: string; data: PurchaseOrder }>(`/purchase/po/${id}`, {
            method: "PUT",
        }),

    deletePurchaseOrder: (id: number) =>
        request<{ message: string }>(`/purchase/po/${id}`, {
            method: "DELETE",
        }),
};
