import { request } from "./client";

export interface Vendor {
    ID: number;
    Name: string;
    Address: string;
    TaxID: string;
    CreatedAt: string;
    UpdatedAt: string;
}

export const vendorApi = {
    getVendors: () => request<Vendor[]>("/purchase/vendor"),

    getVendorById: (id: number) => request<Vendor>(`/purchase/vendor/${id}`),

    createVendor: (data: {
        name: string;
        address: string;
        tax_id: string;
    }) =>
        request<{ message: string; data: Vendor }>("/purchase/vendor", {
            method: "POST",
            body: JSON.stringify(data),
        }),

    updateVendor: (id: number, data: {
        name?: string;
        address?: string;
        tax_id?: string;
    }) =>
        request<{ message: string; data: Vendor }>(`/purchase/vendor/${id}`, {
            method: "PUT",
            body: JSON.stringify(data),
        }),

    deleteVendor: (id: number) =>
        request<{ message: string }>(`/purchase/vendor/${id}`, {
            method: "DELETE",
        }),
};
