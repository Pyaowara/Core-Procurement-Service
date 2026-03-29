import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { poApi, type PurchaseOrder } from "@/lib/api/po";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { HugeiconsIcon } from "@hugeicons/react";
import { Download02Icon, Delete02Icon } from "@hugeicons/core-free-icons";
import ConfirmModal from "@/components/ConfirmModal";
import { toast } from "react-hot-toast";

export default function PoDetailPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { user } = useAuth();
    const [po, setPo] = useState<PurchaseOrder | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [isReceiving, setIsReceiving] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [receiveConfirmOpen, setReceiveConfirmOpen] = useState(false);
    const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
    const canReceiveGoods = user && ["PurchaseOfficer", "Admin"].includes(user.role);

    const load = async () => {
        if (!id) {
            setError("Invalid PO ID");
            setLoading(false);
            return;
        }
        setLoading(true);
        setError("");
        try {
            const data = await poApi.getPurchaseOrderById(Number(id));
            setPo(data);
            console.log("PO loaded:", data);
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to load PO details";
            console.error("Error loading PO:", err);
            setError(errorMsg);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
    }, [id]);

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString("th-TH", {
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
        });
    };

    const formatCurrency = (amount: number) => {
        return new Intl.NumberFormat("th-TH", {
            style: "currency",
            currency: "THB",
        }).format(amount);
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

    const handleReceiveGoods = async () => {
        setIsReceiving(true);
        try {
            await poApi.receiveGoods(Number(id));
            toast.success("Goods received successfully");
            setReceiveConfirmOpen(false);
            load();
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to receive goods";
            console.error("Error receiving goods:", err);
            toast.error(errorMsg);
        } finally {
            setIsReceiving(false);
        }
    };

    const handleDeletePO = async () => {
        setIsDeleting(true);
        try {
            await poApi.deletePurchaseOrder(Number(id));
            toast.success("Purchase Order deleted successfully");
            setDeleteConfirmOpen(false);
            navigate("/po");
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to delete PO";
            console.error("Error deleting PO:", err);
            toast.error(errorMsg);
        } finally {
            setIsDeleting(false);
        }
    };

    if (loading) {
        return (
            <div className="mx-auto max-w-4xl p-6">
                <div className="rounded-md bg-blue-50 p-4 text-sm text-blue-700 border border-blue-200">
                    Loading PO details...
                </div>
            </div>
        );
    }

    if (error || !po) {
        return (
            <div className="mx-auto max-w-4xl p-6">
                <Button onClick={() => navigate("/po")} className="mb-4">
                    ← Back to Purchase Orders
                </Button>
                <div className="rounded-md bg-destructive/10 p-4 text-sm text-destructive border border-destructive/20">
                    <p className="font-medium">Error loading PO details:</p>
                    <p className="text-xs mt-1 font-mono">{error || "PO not found"}</p>
                </div>
            </div>
        );
    }

    if (po.IsDeleted) {
        return (
            <div className="mx-auto max-w-4xl p-6">
                <Button onClick={() => navigate("/po")} className="mb-4">
                    ← Back to Purchase Orders
                </Button>
                <div className="rounded-md bg-destructive/10 p-4 text-sm text-destructive border border-destructive/20">
                    <p className="font-medium">This Purchase Order has been deleted and cannot be edited or modified.</p>
                </div>
            </div>
        );
    }

    const totalAmount = po.Items.reduce((sum, item) => sum + (item.TotalPrice || 0), 0);

    return (
        <div className="mx-auto max-w-5xl p-6">
            <Button onClick={() => navigate("/po")} variant="outline" className="mb-6">
                ← Back to Purchase Orders
            </Button>

            {/* PO Header */}
            <div className="mb-8 rounded-lg border border-border bg-card p-6">
                <div className="mb-4 flex items-start justify-between">
                    <div>
                        <h1 className="text-3xl font-bold text-foreground">{po.PONumber}</h1>
                        <p className="text-sm text-muted-foreground">Purchase Order Details</p>
                    </div>
                    <Badge 
                        variant={getStatusColor(po.Status)} 
                        className={`text-base px-3 py-1 ${po.Status === "SENT" ? "bg-amber-100 text-amber-900 border-amber-300" : ""}`}
                    >
                        {po.Status}
                    </Badge>
                </div>

                <div className="grid grid-cols-2 gap-6 md:grid-cols-3">
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Vendor</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">
                            {po.Vendor?.Name || `Vendor #${po.VendorID}`}
                        </p>
                    </div>
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Purpose</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">{po.Purpose}</p>
                    </div>
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Due Date</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">
                            {formatDate(po.DueDate)}
                        </p>
                    </div>
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Credit Days</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">{po.CreditDay} days</p>
                    </div>
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Created At</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">
                            {formatDate(po.CreatedAt)}
                        </p>
                    </div>
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Last Updated</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">
                            {formatDate(po.UpdatedAt)}
                        </p>
                    </div>
                </div>
            </div>

            {/* PO Items */}
            <div className="mb-8">
                <h2 className="mb-4 text-xl font-bold">Items</h2>
                <div className="rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>SKU</TableHead>
                                <TableHead>Item Name</TableHead>
                                <TableHead>Description</TableHead>
                                <TableHead className="text-right">Qty</TableHead>
                                <TableHead className="text-right">Unit Price</TableHead>
                                <TableHead className="text-right">Discount</TableHead>
                                <TableHead className="text-right">Total</TableHead>
                                <TableHead>Required Date</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {po.Items.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={8} className="text-center text-muted-foreground">
                                        No items in this PO
                                    </TableCell>
                                </TableRow>
                            ) : (
                                po.Items.map((item) => (
                                    <TableRow key={item.ID}>
                                        <TableCell className="font-mono font-semibold">{item.SKU}</TableCell>
                                        <TableCell>{item.ItemName}</TableCell>
                                        <TableCell className="text-sm text-muted-foreground">{item.Description || "—"}</TableCell>
                                        <TableCell className="text-right">{item.Quantity}</TableCell>
                                        <TableCell className="text-right">
                                            {formatCurrency(item.PricePerUnit)}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            {item.Discount > 0 ? (
                                                <span>
                                                    {item.Discount}
                                                    {item.DiscountUnit}
                                                </span>
                                            ) : (
                                                <span className="text-muted-foreground">—</span>
                                            )}
                                        </TableCell>
                                        <TableCell className="text-right font-semibold">
                                            {formatCurrency(item.TotalPrice)}
                                        </TableCell>
                                        <TableCell className="text-muted-foreground">
                                            {formatDate(item.RequiredDate)}
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </div>
            </div>

            {/* Summary */}
            <div className="rounded-lg border border-border bg-card p-6">
                <div className="flex justify-end">
                    <div className="w-full max-w-xs space-y-2">
                        <div className="flex justify-between text-sm">
                            <span className="text-muted-foreground">Total Items:</span>
                            <span className="font-semibold">{po.Items.length}</span>
                        </div>
                        <div className="flex justify-between border-t pt-2 text-lg font-bold">
                            <span>Total Amount:</span>
                            <span>{formatCurrency(totalAmount)}</span>
                        </div>
                    </div>
                </div>
            </div>

            {/* Action Buttons */}
            {!po.IsDeleted && (po.Status === "SENT" || po.Status === "FAILED") && canReceiveGoods && (
                <div className="mt-6 flex items-center justify-end gap-2">
                    <Button
                        onClick={() => setDeleteConfirmOpen(true)}
                        variant="destructive"
                        disabled={isReceiving || isDeleting}
                        className="gap-2"
                    >
                        <HugeiconsIcon
                            icon={Delete02Icon}
                            strokeWidth={2}
                            className="w-4 h-4"
                        />
                        {isDeleting ? "Deleting..." : "Delete"}
                    </Button>
                    <Button
                        onClick={() => setReceiveConfirmOpen(true)}
                        variant="default"
                        disabled={isReceiving || isDeleting}
                        className="gap-2 bg-green-600 hover:bg-green-700"
                    >
                        <HugeiconsIcon
                            icon={Download02Icon}
                            strokeWidth={2}
                            className="w-4 h-4"
                        />
                        {isReceiving ? "Receiving..." : "Receive Goods"}
                    </Button>
                </div>
            )}

            {/* Receive Goods Confirmation Modal */}
            <ConfirmModal
                open={receiveConfirmOpen}
                onOpenChange={setReceiveConfirmOpen}
                title="Confirm Receive Goods"
                description="Are you sure you want to mark all items in this Purchase Order as received? The PO status will be updated to COMPLETED."
                onConfirm={handleReceiveGoods}
                isLoading={isReceiving}
                confirmText="Receive"
                cancelText="Cancel"
                variant="default"
            />

            {/* Delete Confirmation Modal */}
            <ConfirmModal
                open={deleteConfirmOpen}
                onOpenChange={setDeleteConfirmOpen}
                title="Confirm Delete"
                description="Are you sure you want to delete this Purchase Order? This action cannot be undone. The PO will be permanently deleted."
                onConfirm={handleDeletePO}
                isLoading={isDeleting}
                confirmText="Delete"
                cancelText="Cancel"
                variant="destructive"
            />
        </div>
    );
}
