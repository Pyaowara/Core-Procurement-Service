import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { poApi, type PurchaseOrder } from "@/lib/api/po";
import { Button } from "@/components/ui/button";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function PoListPage() {
    const navigate = useNavigate();
    const [pos, setPos] = useState<PurchaseOrder[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [statusFilter, setStatusFilter] = useState<string>("ALL");

    const load = async () => {
        setLoading(true);
        setError("");
        try {
            const data = await poApi.getPurchaseOrders();
            setPos(data || []);
            console.log("Purchase orders loaded:", data);
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to load purchase orders";
            console.error("Error loading POs:", err);
            setError(errorMsg);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, []);

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString("th-TH", {
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
        });
    };

    const getStatusColor = (status: string): "secondary" | "outline" | "default" | "destructive" => {
        switch (status) {
            case "SENT":
                return "outline";
            case "COMPLETED":
                return "default";
            case "FAILED":
                return "destructive";
            default:
                return "secondary";
        }
    };

    // Filter POs based on selected status
    const filteredPos = (statusFilter === "ALL" 
        ? pos 
        : pos.filter(po => po.Status === statusFilter)
    ).sort((a, b) => new Date(b.CreatedAt).getTime() - new Date(a.CreatedAt).getTime());

    const statusOptions = ["ALL", "SENT", "COMPLETED", "FAILED"];

    return (
        <div className="mx-auto max-w-6xl p-6">
            <div className="mb-6 flex items-center justify-between">
                <h1 className="text-2xl font-bold">Purchase Orders</h1>
                <div className="flex gap-2">
                    <Button onClick={load} disabled={loading}>
                        {loading ? "Loading..." : "Refresh"}
                    </Button>
                </div>
            </div>

            {/* Status Filter */}
            <div className="mb-6 flex flex-wrap gap-2">
                {statusOptions.map((status) => (
                    <Button
                        key={status}
                        variant={statusFilter === status ? "default" : "outline"}
                        size="sm"
                        onClick={() => setStatusFilter(status)}
                    >
                        {status === "ALL" ? "All Status" : status}
                    </Button>
                ))}
            </div>

            {error && (
                <div className="mb-4 rounded-md bg-destructive/10 p-4 text-sm text-destructive border border-destructive/20">
                    <p className="font-medium">Error loading purchase orders:</p>
                    <p className="text-xs mt-1 font-mono">{error}</p>
                </div>
            )}

            {loading && (
                <div className="mb-4 rounded-md bg-blue-50 p-4 text-sm text-blue-700 border border-blue-200">
                    Loading purchase orders...
                </div>
            )}

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>PO Number</TableHead>
                            <TableHead>Vendor</TableHead>
                            <TableHead>Purpose</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Due Date</TableHead>
                            <TableHead>Created At</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {filteredPos.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center text-muted-foreground">
                                    {pos.length === 0 ? "No purchase orders yet" : `No ${statusFilter === "ALL" ? "" : statusFilter + " "}purchase orders`}
                                </TableCell>
                            </TableRow>
                        ) : (
                            filteredPos.map((po) => (
                                <TableRow 
                                    key={po.ID} 
                                    className={`cursor-pointer hover:bg-muted/50 transition-colors ${po.IsDeleted ? "opacity-50" : ""}`}
                                    onClick={() => navigate(`/po/${po.ID}`)}
                                >
                                    <TableCell className="font-mono font-semibold">
                                        {po.PONumber}
                                        {po.IsDeleted && <span className="ml-2 text-sm text-destructive">(Deleted)</span>}
                                    </TableCell>
                                    <TableCell>{po.Vendor?.Name || `Vendor #${po.VendorID}`}</TableCell>
                                    <TableCell>{po.Purpose}</TableCell>
                                    <TableCell>
                                        <Badge 
                                            variant={getStatusColor(po.Status)}
                                            className={po.Status === "SENT" ? "bg-amber-100 text-amber-900 border-amber-300" : ""}
                                        >
                                            {po.Status}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="text-muted-foreground">
                                        {formatDate(po.DueDate)}
                                    </TableCell>
                                    <TableCell className="text-muted-foreground">
                                        {formatDate(po.CreatedAt)}
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
