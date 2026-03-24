import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { prApi, type PurchaseRequest } from "@/lib/api/pr";
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
import { SentIcon, Delete02Icon } from "@hugeicons/core-free-icons";
import PrForm from "@/components/PrForm";
import ConfirmModal from "@/components/ConfirmModal";
import { toast } from "sonner";

export default function PrDetailPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [pr, setPr] = useState<PurchaseRequest | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [isEditFormOpen, setIsEditFormOpen] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [submitConfirmOpen, setSubmitConfirmOpen] = useState(false);
    const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

    const load = async () => {
        if (!id) {
            setError("Invalid PR ID");
            setLoading(false);
            return;
        }
        setLoading(true);
        setError("");
        try {
            const data = await prApi.getPurchaseRequestById(Number(id));
            setPr(data);
            console.log("PR loaded:", data);
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to load PR details";
            console.error("Error loading PR:", err);
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
            case "DRAFT":
                return "secondary";
            case "PENDING":
                return "outline";
            case "APPROVED":
                return "default";
            case "REJECTED":
                return "destructive";
            default:
                return "secondary";
        }
    };

    const handleSubmitPR = async () => {
        setIsSubmitting(true);
        try {
            await prApi.submitPurchaseRequest(Number(id));
            toast.success("Purchase Request submitted successfully");
            setSubmitConfirmOpen(false);
            load();
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to submit PR";
            console.error("Error submitting PR:", err);
            toast.error(errorMsg);
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleDeletePR = async () => {
        setIsDeleting(true);
        try {
            await prApi.deletePurchaseRequest(Number(id));
            toast.success("Purchase Request deleted successfully");
            setDeleteConfirmOpen(false);
            navigate("/pr");
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to delete PR";
            console.error("Error deleting PR:", err);
            toast.error(errorMsg);
        } finally {
            setIsDeleting(false);
        }
    };

    if (loading) {
        return (
            <div className="mx-auto max-w-4xl p-6">
                <div className="rounded-md bg-blue-50 p-4 text-sm text-blue-700 border border-blue-200">
                    Loading PR details...
                </div>
            </div>
        );
    }

    if (error || !pr) {
        return (
            <div className="mx-auto max-w-4xl p-6">
                <Button onClick={() => navigate("/pr")} className="mb-4">
                    ← Back to Purchase Requests
                </Button>
                <div className="rounded-md bg-destructive/10 p-4 text-sm text-destructive border border-destructive/20">
                    <p className="font-medium">Error loading PR details:</p>
                    <p className="text-xs mt-1 font-mono">{error || "PR not found"}</p>
                </div>
            </div>
        );
    }

    if (pr.isDeleted) {
        return (
            <div className="mx-auto max-w-4xl p-6">
                <Button onClick={() => navigate("/pr")} className="mb-4">
                    ← Back to Purchase Requests
                </Button>
                <div className="rounded-md bg-destructive/10 p-4 text-sm text-destructive border border-destructive/20">
                    <p className="font-medium">This Purchase Request has been deleted and cannot be edited or modified.</p>
                </div>
            </div>
        );
    }

    const totalAmount = pr.Items.reduce((sum, item) => sum + (item.TotalPrice || 0), 0);

    return (
        <div className="mx-auto max-w-5xl p-6">
            <Button onClick={() => navigate("/pr")} variant="outline" className="mb-6">
                ← Back to Purchase Requests
            </Button>

            {/* PR Header */}
            <div className="mb-8 rounded-lg border border-border bg-card p-6">
                <div className="mb-4 flex items-start justify-between">
                    <div>
                        <h1 className="text-3xl font-bold text-foreground">{pr.PRNumber}</h1>
                        <p className="text-sm text-muted-foreground">Purchase Request Details</p>
                    </div>
                    <div className="flex items-center gap-2">
                        {pr.Status === "DRAFT" && (
                            <Button
                                onClick={() => setIsEditFormOpen(true)}
                                variant="default"
                                disabled={isSubmitting || isDeleting}
                            >
                                Edit
                            </Button>
                        )}
                        <Badge variant={getStatusColor(pr.Status)} className="text-base px-3 py-1">
                            {pr.Status}
                        </Badge>
                    </div>
                </div>

                <div className="grid grid-cols-2 gap-6 md:grid-cols-3">
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Department</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">{pr.Department}</p>
                    </div>
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Purpose</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">{pr.Purpose}</p>
                    </div>
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Created At</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">
                            {formatDate(pr.CreatedAt)}
                        </p>
                    </div>
                    <div>
                        <p className="text-xs font-medium text-muted-foreground">Last Updated</p>
                        <p className="mt-1 text-sm font-semibold text-foreground">
                            {formatDate(pr.UpdatedAt)}
                        </p>
                    </div>
                </div>
            </div>

            {/* PR Items */}
            <div className="mb-8">
                <h2 className="mb-4 text-xl font-bold">Items</h2>
                <div className="rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>SKU</TableHead>
                                <TableHead>Item Name</TableHead>
                                <TableHead className="text-right">Qty</TableHead>
                                <TableHead className="text-right">Unit Price</TableHead>
                                <TableHead className="text-right">Discount</TableHead>
                                <TableHead className="text-right">Total</TableHead>
                                <TableHead>Required Date</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {pr.Items.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center text-muted-foreground">
                                        No items in this PR
                                    </TableCell>
                                </TableRow>
                            ) : (
                                pr.Items.map((item) => (
                                    <TableRow key={item.ID}>
                                        <TableCell className="font-mono font-semibold">{item.SKU}</TableCell>
                                        <TableCell>{item.ItemName}</TableCell>
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
                            <span className="font-semibold">{pr.Items.length}</span>
                        </div>
                        <div className="flex justify-between border-t pt-2 text-lg font-bold">
                            <span>Total Amount:</span>
                            <span>{formatCurrency(totalAmount)}</span>
                        </div>
                    </div>
                </div>
            </div>

            {/* Action Buttons */}
            {!pr.isDeleted && pr.Status === "DRAFT" && (
                <div className="mt-6 flex items-center justify-end gap-2">
                    <Button
                        onClick={() => setDeleteConfirmOpen(true)}
                        variant="destructive"
                        disabled={isSubmitting || isDeleting}
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
                        onClick={() => setSubmitConfirmOpen(true)}
                        variant="default"
                        disabled={isSubmitting || isDeleting}
                        className="gap-2 bg-green-600 hover:bg-green-700"
                    >
                        <HugeiconsIcon
                            icon={SentIcon}
                            strokeWidth={2}
                            className="w-4 h-4"
                        />
                        {isSubmitting ? "Submitting..." : "Submit"}
                    </Button>
                    
                </div>
            )}

            {/* Edit PR Form */}
            {pr && (
                <PrForm
                    open={isEditFormOpen}
                    onOpenChange={setIsEditFormOpen}
                    mode="edit"
                    existingPr={pr}
                    onSuccess={load}
                />
            )}

            {/* Submit Confirmation Modal */}
            <ConfirmModal
                open={submitConfirmOpen}
                onOpenChange={setSubmitConfirmOpen}
                title="Confirm Submit"
                description="Are you sure you want to submit this Purchase Request? This action cannot be undone. The PR will be sent for approval."
                onConfirm={handleSubmitPR}
                isLoading={isSubmitting}
                confirmText="Submit"
                cancelText="Cancel"
                variant="default"
            />

            {/* Delete Confirmation Modal */}
            <ConfirmModal
                open={deleteConfirmOpen}
                onOpenChange={setDeleteConfirmOpen}
                title="Confirm Delete"
                description="Are you sure you want to delete this Purchase Request? This action cannot be undone. The PR will be permanently deleted."
                onConfirm={handleDeletePR}
                isLoading={isDeleting}
                confirmText="Delete"
                cancelText="Cancel"
                variant="destructive"
            />
        </div>
    );
}
