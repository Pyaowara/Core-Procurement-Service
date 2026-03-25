import { createContext, useContext, useState, type ReactNode } from "react";
import { vendorApi, type Vendor } from "@/lib/api/vendor";

interface VendorContextType {
    vendors: Vendor[];
    loading: boolean;
    refresh: () => Promise<void>;
}

const VendorContext = createContext<VendorContextType | null>(null);

export function VendorProvider({ children }: { children: ReactNode }) {
    const [vendors, setVendors] = useState<Vendor[]>([]);
    const [loading, setLoading] = useState(false);

    const refresh = async () => {
        setLoading(true);
        try {
            const data = await vendorApi.getVendors();
            setVendors(data || []);
        } catch (error) {
            console.error("Failed to fetch vendors:", error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <VendorContext.Provider value={{ vendors, loading, refresh }}>
            {children}
        </VendorContext.Provider>
    );
}

export function useVendor() {
    const ctx = useContext(VendorContext);
    if (!ctx) throw new Error("useVendor must be used within VendorProvider");
    return ctx;
}
