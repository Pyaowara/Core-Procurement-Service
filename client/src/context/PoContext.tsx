import { createContext, useContext, useState, type ReactNode } from "react";
import { poApi, type PurchaseOrder } from "@/lib/api/po";

interface PoContextType {
    pos: PurchaseOrder[];
    loading: boolean;
    refresh: () => Promise<void>;
}

const PoContext = createContext<PoContextType | null>(null);

export function PoProvider({ children }: { children: ReactNode }) {
    const [pos, setPos] = useState<PurchaseOrder[]>([]);
    const [loading, setLoading] = useState(false);

    const refresh = async () => {
        setLoading(true);
        try {
            const data = await poApi.getPurchaseOrders();
            setPos(data || []);
        } catch (error) {
            console.error("Failed to fetch purchase orders:", error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <PoContext.Provider value={{ pos, loading, refresh }}>
            {children}
        </PoContext.Provider>
    );
}

export function usePo() {
    const ctx = useContext(PoContext);
    if (!ctx) throw new Error("usePo must be used within PoProvider");
    return ctx;
}
