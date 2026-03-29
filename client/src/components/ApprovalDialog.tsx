import { useState, useEffect } from "react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

interface ApprovalDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    action: "approve" | "reject";
    userRole: string;
    isLoading: boolean;
    onConfirm: (reason: string) => Promise<void>;
}

export default function ApprovalDialog({
    open,
    onOpenChange,
    action,
    userRole,
    isLoading,
    onConfirm,
}: ApprovalDialogProps) {
    const [reason, setReason] = useState("");

    // Clear reason when dialog closes
    useEffect(() => {
        if (!open) {
            setReason("");
        }
    }, [open]);

    const actionLabel = action === "approve" ? "approve" : "reject";
    const actionLabelCapitalized = action === "approve" ? "Approve" : "Reject";

    const handleConfirm = async () => {
        await onConfirm(reason);
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle>Confirm {actionLabelCapitalized}</DialogTitle>
                    <DialogDescription>
                        Are you sure you want to {actionLabel} this Purchase Request as{" "}
                        <span className="font-semibold text-foreground">{userRole}</span>?
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-3 py-4">
                    <div className="space-y-2">
                        <Label htmlFor="reason" className="text-sm font-medium">
                            Reason
                        </Label>
                        <Textarea
                            id="reason"
                            placeholder="Enter your reason (optional)..."
                            value={reason}
                            onChange={(e) => setReason(e.target.value)}
                            disabled={isLoading}
                        />
                        <p className="text-xs text-muted-foreground">
                            If left blank, will auto-comment with your user ID
                        </p>
                    </div>
                </div>

                <DialogFooter>
                    <Button
                        variant="outline"
                        onClick={() => onOpenChange(false)}
                        disabled={isLoading}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleConfirm}
                        disabled={isLoading}
                        className={
                            action === "reject"
                                ? "bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                : "bg-green-600 hover:bg-green-700"
                        }
                    >
                        {isLoading
                            ? `${actionLabelCapitalized}ing...`
                            : `Yes, ${actionLabel}`}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
