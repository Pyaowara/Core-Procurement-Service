import { useAuth } from "@/context/AuthContext";
import { approvalApi, type ApprovalInstance } from "@/lib/api/approval";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { HugeiconsIcon } from "@hugeicons/react";
import { ThumbsUpIcon, ThumbsDownIcon, ArrowDown01Icon, ArrowUp01Icon } from "@hugeicons/core-free-icons";
import ApprovalDialog from "./ApprovalDialog";
import { useState } from "react";
import { toast } from "react-hot-toast";

interface ApprovalStatusDisplayProps {
    approval: ApprovalInstance;
    onActionComplete: () => void;
}

export default function ApprovalStatusDisplay({
    approval,
    onActionComplete,
}: ApprovalStatusDisplayProps) {
    const { user } = useAuth();
    const workflowId = (approval as any).WorkflowID;
    const [approveDialogOpen, setApproveDialogOpen] = useState(false);
    const [rejectDialogOpen, setRejectDialogOpen] = useState(false);
    const [isApproving, setIsApproving] = useState(false);
    const [isRejecting, setIsRejecting] = useState(false);
    const [isCollapsed, setIsCollapsed] = useState(true);

    const latestRejectedAction = [...approval.Actions]
        .filter((action) => action.ActionType === "REJECTED")
        .sort((a, b) => new Date(b.CreatedAt).getTime() - new Date(a.CreatedAt).getTime())[0];
    const rejectedReason = latestRejectedAction?.Comment?.trim();

    // Get the current step
    const currentStep = approval.Steps.find(
        (s) => s.StepOrder === approval.CurrentStep
    );

    // Check if user can approve/reject (role must match current step's required role)
    const canApproveOrReject =
        user && currentStep && verifyApprovalRole(user.role, currentStep.Role);

    // Map user roles to approval roles
    function verifyApprovalRole(userRole: string, requiredApprovalRole: string): boolean {
        const roleMapping: Record<string, string[]> = {
            "Employee": ["Employee", "Manager", "PurchaseOfficer", "Executive", "Admin"],
            "Manager": ["Manager", "Executive", "Admin"],
            "PurchaseOfficer": ["PurchaseOfficer", "Executive", "Admin"],
            "EXECUTIVE": ["Executive", "Admin"],
        };

        const allowedRoles = roleMapping[requiredApprovalRole] || [];
        return allowedRoles.includes(userRole);
    }

    const getStepStatusIcon = (status: string) => {
        switch (status) {
            case "APPROVED":
                return (
                    <div className="w-5 h-5 rounded-full bg-green-100 flex items-center justify-center">
                        <span className="text-green-600 font-bold text-xs">✓</span>
                    </div>
                );
            case "REJECTED":
                return (
                    <div className="w-5 h-5 rounded-full bg-red-100 flex items-center justify-center">
                        <span className="text-red-600 font-bold text-xs">✕</span>
                    </div>
                );
            case "PENDING":
                return (
                    <div className="w-5 h-5 rounded-full bg-amber-100 flex items-center justify-center">
                        <span className="text-amber-600 text-xs">⏱</span>
                    </div>
                );
            default:
                return null;
        }
    };

    const getStepStatusBadge = (status: string): "secondary" | "outline" | "default" | "destructive" => {
        switch (status) {
            case "APPROVED":
                return "default";
            case "REJECTED":
                return "destructive";
            case "PENDING":
                return "outline";
            default:
                return "secondary";
        }
    };

    const getStepDisplayStatus = (stepStatus: string, stepOrder: number) => {
        if (
            approval.Status === "REJECTED" &&
            stepStatus === "PENDING" &&
            stepOrder > approval.CurrentStep
        ) {
            return "SKIPPED";
        }
        return stepStatus;
    };

    const getStepSubtext = (stepStatus: string, stepOrder: number, actionAt?: string | null) => {
        if (stepStatus === "PENDING" && approval.Status === "PENDING" && stepOrder === approval.CurrentStep) {
            return "Currently awaiting approval";
        }
        if (approval.Status === "REJECTED" && stepStatus === "PENDING" && stepOrder > approval.CurrentStep) {
            return "Not required after rejection";
        }
        if (stepStatus === "PENDING") {
            return "Pending";
        }
        if (stepStatus === "APPROVED") {
            return `Approved on ${new Date(actionAt ?? "").toLocaleDateString()}`;
        }
        return `Rejected on ${new Date(actionAt ?? "").toLocaleDateString()}`;
    };

    const getRoleDisplayName = (role: string): string => {
        const roleNames: Record<string, string> = {
            "Employee": "Employee",
            "Manager": "Department Manager",
            "PurchaseOfficer": "Purchase Officer",
            "EXECUTIVE": "Executive",
        };
        return roleNames[role] || role;
    };

    const handleApprove = async (reason: string) => {
        setIsApproving(true);
        try {
            if (!workflowId) {
                throw new Error("Workflow ID not found");
            }
            // Use provided reason or generate default with role
            const finalReason = reason.trim() || `${user?.role} Approved`;
            await approvalApi.approveStepByWorkflow(workflowId, finalReason);
            toast.success("Purchase Request approved successfully");
            setApproveDialogOpen(false);
            onActionComplete();
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to approve PR";
            console.error("Error approving PR:", err);
            toast.error(errorMsg);
        } finally {
            setIsApproving(false);
        }
    };

    const handleReject = async (reason: string) => {
        setIsRejecting(true);
        try {
            if (!workflowId) {
                throw new Error("Workflow ID not found");
            }
            // Use provided reason or generate default with role
            const finalReason = reason.trim() || `${user?.role} Rejected`;
            await approvalApi.rejectStepByWorkflow(workflowId, finalReason);
            toast.success("Purchase Request rejected successfully");
            setRejectDialogOpen(false);
            onActionComplete();
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to reject PR";
            console.error("Error rejecting PR:", err);
            toast.error(errorMsg);
        } finally {
            setIsRejecting(false);
        }
    };


    return (
        <>
            <Card className="mb-8">
                <CardHeader>
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                        <div>
                            <CardTitle className="text-lg">Approval Status</CardTitle>
                            <CardDescription>
                                Current Status: <span className="font-semibold">{approval.Status}</span>
                            </CardDescription>
                            {approval.Status === "REJECTED" && rejectedReason && (
                                <p className="mt-1 text-sm text-destructive">
                                    Reject Reason: <span className="font-medium">{rejectedReason}</span>
                                </p>
                            )}
                        </div>

                        <div className="flex flex-wrap items-center gap-2">
                            {approval.Status === "PENDING" && canApproveOrReject && currentStep && (
                                <>
                                    <Button
                                        onClick={() => setApproveDialogOpen(true)}
                                        disabled={isApproving || isRejecting}
                                        className="gap-2 bg-green-600 hover:bg-green-700"
                                    >
                                        <HugeiconsIcon
                                            icon={ThumbsUpIcon}
                                            strokeWidth={2}
                                            className="w-4 h-4"
                                        />
                                        {isApproving ? "Approving..." : "Approve"}
                                    </Button>
                                    <Button
                                        onClick={() => setRejectDialogOpen(true)}
                                        disabled={isApproving || isRejecting}
                                        variant="destructive"
                                        className="gap-2"
                                    >
                                        <HugeiconsIcon
                                            icon={ThumbsDownIcon}
                                            strokeWidth={2}
                                            className="w-4 h-4"
                                        />
                                        {isRejecting ? "Rejecting..." : "Reject"}
                                    </Button>
                                </>
                            )}

                            <Button
                                variant="outline"
                                onClick={() => setIsCollapsed((prev) => !prev)}
                                className="gap-2"
                            >
                                <HugeiconsIcon
                                    icon={isCollapsed ? ArrowDown01Icon : ArrowUp01Icon}
                                    strokeWidth={2}
                                    className="w-4 h-4"
                                />
                                {isCollapsed ? "Show Steps" : "Hide Steps"}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                {!isCollapsed && (
                    <CardContent className="space-y-6">
                    {/* Approval Steps Timeline */}
                    <div className="space-y-4">
                        <h3 className="text-sm font-semibold">Approval Steps</h3>
                        {approval.Steps.map((step, index) => (
                            <div key={step.ID} className="flex items-start gap-4">
                                {/* Timeline connector */}
                                <div className="flex flex-col items-center">
                                    <div className="rounded-full bg-background border-2 border-border p-2">
                                        {getStepStatusIcon(step.Status)}
                                    </div>
                                    {index < approval.Steps.length - 1 && (
                                        <div
                                            className={`w-0.5 h-12 my-2 ${
                                                step.Status === "PENDING"
                                                    ? "bg-muted"
                                                    : "bg-green-600"
                                            }`}
                                        />
                                    )}
                                </div>

                                {/* Step Details */}
                                <div className="flex-1 pt-2">
                                    <div className="flex items-center justify-between mb-1">
                                        <div>
                                            <p className="font-medium text-sm">
                                                Step {step.StepOrder}: {getRoleDisplayName(step.Role)}
                                            </p>
                                            <p className="text-xs text-muted-foreground">
                                                {getStepSubtext(step.Status, step.StepOrder, step.ActionAt)}
                                            </p>
                                        </div>
                                        <Badge
                                            variant={getStepDisplayStatus(step.Status, step.StepOrder) === "SKIPPED" ? "secondary" : getStepStatusBadge(step.Status)}
                                        >
                                            {getStepDisplayStatus(step.Status, step.StepOrder)}
                                        </Badge>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                    </CardContent>
                )}
            </Card>

            {/* Approval Confirmation Dialog */}
            <ApprovalDialog
                open={approveDialogOpen}
                onOpenChange={setApproveDialogOpen}
                action="approve"
                userRole={user?.role || ""}
                isLoading={isApproving}
                onConfirm={handleApprove}
            />

            {/* Rejection Confirmation Dialog */}
            <ApprovalDialog
                open={rejectDialogOpen}
                onOpenChange={setRejectDialogOpen}
                action="reject"
                userRole={user?.role || ""}
                isLoading={isRejecting}
                onConfirm={handleReject}
            />
        </>
    );
}
