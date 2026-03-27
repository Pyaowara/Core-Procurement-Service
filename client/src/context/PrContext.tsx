import { createContext, useContext, useState, type ReactNode } from "react";
import { prApi, type PurchaseRequest } from "@/lib/api/pr";

interface PrContextType {
    prs: PurchaseRequest[];
    loading: boolean;
    refresh: () => Promise<void>;
}

const PrContext = createContext<PrContextType | null>(null);

export function PrProvider({ children }: { children: ReactNode }) {
    const [prs, setPrs] = useState<PurchaseRequest[]>([]);
    const [loading, setLoading] = useState(false);

    const refresh = async () => {
        setLoading(true);
        try {
            const data = await prApi.getPurchaseRequests();
            setPrs(data || []);
        } catch (error) {
            console.error("Failed to fetch purchase requests:", error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <PrContext.Provider value={{ prs, loading, refresh }}>
            {children}
        </PrContext.Provider>
    );
}

export function usePr() {
    const ctx = useContext(PrContext);
    if (!ctx) throw new Error("usePr must be used within PrProvider");
    return ctx;
}
