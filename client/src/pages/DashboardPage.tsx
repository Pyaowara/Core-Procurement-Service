import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import { prApi, type PurchaseRequest } from "@/lib/api/pr";
import { poApi, type PurchaseOrder } from "@/lib/api/po";
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
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";

type DashboardApprovalItem = {
    pr: PurchaseRequest;
    approval: ApprovalInstance;
};

function normalizeRole(role: string): string {
    const trimmed = role.trim();
    const upper = trimmed.toUpperCase();
    if (upper === "EXECUTIVE") return "Executive";
    if (upper === "PURCHASEOFFICER") return "PurchaseOfficer";
    if (upper === "MANAGER") return "Manager";
    if (upper === "EMPLOYEE") return "Employee";
    if (upper === "ADMIN") return "Admin";
    return trimmed;
}

function canUserApproveStep(userRole: string, stepRole: string): boolean {
    const normalizedUserRole = normalizeRole(userRole);
    const normalizedStepRole = normalizeRole(stepRole);
    return normalizedUserRole === normalizedStepRole;
}

function formatDate(value: string): string {
    return new Date(value).toLocaleDateString("th-TH", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
    });
}

export default function DashboardPage() {
    const { user } = useAuth();
    const [prs, setPrs] = useState<PurchaseRequest[]>([]);
    const [pos, setPos] = useState<PurchaseOrder[]>([]);
    const [pendingApprovals, setPendingApprovals] = useState<DashboardApprovalItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

    const canViewPO = user ? ["PurchaseOfficer", "Admin"].includes(user.role) : false;

    const loadDashboard = async () => {
        setLoading(true);
        setError("");

        try {
            const [prList, poList] = await Promise.all([
                prApi.getPurchaseRequests(),
                canViewPO ? poApi.getPurchaseOrders() : Promise.resolve([]),
            ]);

            const normalizedPrs = prList || [];
            setPrs(normalizedPrs);
            setPos((poList as PurchaseOrder[]) || []);

            if (!user) {
                setPendingApprovals([]);
                return;
            }

            const pendingPrs = normalizedPrs.filter(
                (pr) => pr.Status === "PENDING" && !!pr.WorkflowID && !pr.IsDeleted
            );

            const approvalResults = await Promise.all(
                pendingPrs.map(async (pr) => {
                    try {
                        const approval = await approvalApi.getApprovalByWorkflow(pr.WorkflowID);
                        return { pr, approval };
                    } catch {
                        return null;
                    }
                })
            );

            const requiresMyApproval = approvalResults
                .filter((item): item is DashboardApprovalItem => item !== null)
                .filter((item) => {
                    if (item.approval.Status !== "PENDING") return false;
                    const currentStep = item.approval.Steps.find(
                        (step) => step.StepOrder === item.approval.CurrentStep
                    );
                    if (!currentStep) return false;
                    return canUserApproveStep(user.role, currentStep.Role);
                })
                .sort((a, b) => {
                    const left = new Date(a.pr.UpdatedAt).getTime();
                    const right = new Date(b.pr.UpdatedAt).getTime();
                    return right - left;
                });

            setPendingApprovals(requiresMyApproval);
        } catch (err: unknown) {
            const message = err instanceof Error ? err.message : "Failed to load dashboard data";
            setError(message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadDashboard();
    }, [user?.role]);

    const metrics = useMemo(() => {
        const draft = prs.filter((pr) => pr.Status === "DRAFT").length;
        const pending = prs.filter((pr) => pr.Status === "PENDING").length;
        const approved = prs.filter((pr) => pr.Status === "APPROVED").length;
        const rejected = prs.filter((pr) => pr.Status === "REJECTED").length;

        const poDraft = pos.filter((po) => po.Status === "DRAFT").length;
        const poSent = pos.filter((po) => po.Status === "SENT").length;
        const poCompleted = pos.filter((po) => po.Status === "COMPLETED").length;

        return {
            prTotal: prs.length,
            draft,
            pending,
            approved,
            rejected,
            poTotal: pos.length,
            poDraft,
            poSent,
            poCompleted,
        };
    }, [prs, pos]);

    const roleHeadline = useMemo(() => {
        const normalized = normalizeRole(user?.role || "");
        if (normalized === "Admin") return "Administration Overview";
        if (normalized === "PurchaseOfficer") return "Procurement Operations";
        if (normalized === "Manager") return "Manager Approval Dashboard";
        if (normalized === "Executive") return "Executive Approval Dashboard";
        return "Employee Request Dashboard";
    }, [user?.role]);

    return (
        <div className="mx-auto max-w-6xl p-6">
            <div className="mb-6 flex flex-wrap items-end justify-between gap-3">
                <div>
                    <h1 className="text-3xl font-bold">{roleHeadline}</h1>
                    <p className="text-sm text-muted-foreground">
                        Signed in as {user?.first_name} {user?.last_name} ({user?.role})
                    </p>
                </div>
                <Button onClick={loadDashboard} disabled={loading}>
                    {loading ? "Refreshing..." : "Refresh Dashboard"}
                </Button>
            </div>

            {error && (
                <div className="mb-6 rounded-md border border-destructive/20 bg-destructive/10 p-4 text-sm text-destructive">
                    {error}
                </div>
            )}

            <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Total PR</CardDescription>
                        <CardTitle className="text-2xl font-bold">{metrics.prTotal}</CardTitle>
                    </CardHeader>
                    <CardContent className="text-xs text-muted-foreground">
                        Requests visible to your role
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Pending PR</CardDescription>
                        <CardTitle className="text-2xl font-bold">{metrics.pending}</CardTitle>
                    </CardHeader>
                    <CardContent className="text-xs text-muted-foreground">
                        Waiting for workflow completion
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Approved PR</CardDescription>
                        <CardTitle className="text-2xl font-bold">{metrics.approved}</CardTitle>
                    </CardHeader>
                    <CardContent className="text-xs text-muted-foreground">
                        Ready for next operational step
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Needs Your Approval</CardDescription>
                        <CardTitle className="text-2xl font-bold">{pendingApprovals.length}</CardTitle>
                    </CardHeader>
                    <CardContent className="text-xs text-muted-foreground">
                        Workflows matching your current role
                    </CardContent>
                </Card>
            </div>

            <div className="mb-6 grid gap-4 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle>PR Status Summary</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-wrap gap-2">
                        <Badge variant="secondary">Draft: {metrics.draft}</Badge>
                        <Badge variant="outline">Pending: {metrics.pending}</Badge>
                        <Badge>Approved: {metrics.approved}</Badge>
                        <Badge variant="destructive">Rejected: {metrics.rejected}</Badge>
                    </CardContent>
                </Card>

                {canViewPO && (
                    <Card>
                        <CardHeader>
                            <CardTitle>PO Status Summary</CardTitle>
                        </CardHeader>
                        <CardContent className="flex flex-wrap gap-2">
                            <Badge variant="secondary">Draft: {metrics.poDraft}</Badge>
                            <Badge variant="outline">Sent: {metrics.poSent}</Badge>
                            <Badge>Completed: {metrics.poCompleted}</Badge>
                            <Badge variant="secondary">Total: {metrics.poTotal}</Badge>
                        </CardContent>
                    </Card>
                )}
            </div>

            <Card className="mb-6">
                <CardHeader>
                    <CardTitle>Approval Work Queue</CardTitle>
                    <CardDescription>
                        Purchase requests that are currently waiting for your approval role.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>PR Number</TableHead>
                                    <TableHead>Department</TableHead>
                                    <TableHead>Current Step</TableHead>
                                    <TableHead>Last Update</TableHead>
                                    <TableHead className="text-right">Action</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {pendingApprovals.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={5} className="text-center text-muted-foreground">
                                            No approval tasks for your role right now.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    pendingApprovals.map((item) => {
                                        const currentStep = item.approval.Steps.find(
                                            (step) => step.StepOrder === item.approval.CurrentStep
                                        );
                                        return (
                                            <TableRow key={item.pr.ID}>
                                                <TableCell className="font-mono font-medium">
                                                    {item.pr.PRNumber}
                                                </TableCell>
                                                <TableCell>{item.pr.Department}</TableCell>
                                                <TableCell>
                                                    Step {item.approval.CurrentStep}: {normalizeRole(currentStep?.Role || "Unknown")}
                                                </TableCell>
                                                <TableCell>{formatDate(item.pr.UpdatedAt)}</TableCell>
                                                <TableCell className="text-right">
                                                    <Button asChild size="sm">
                                                        <Link to={`/pr/${item.pr.ID}`}>Review</Link>
                                                    </Button>
                                                </TableCell>
                                            </TableRow>
                                        );
                                    })
                                )}
                            </TableBody>
                        </Table>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
