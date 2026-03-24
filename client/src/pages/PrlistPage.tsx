import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { prApi, type PurchaseRequest } from "@/lib/api/pr";
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
import PrForm from "@/components/PrForm";
import { HugeiconsIcon } from "@hugeicons/react";
import { Add01Icon} from "@hugeicons/core-free-icons";

export default function PrListPage() {
    const navigate = useNavigate();
    const [prs, setPrs] = useState<PurchaseRequest[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [statusFilter, setStatusFilter] = useState<string>("ALL");
    const [prFormOpen, setPrFormOpen] = useState(false);

    const load = async () => {
        setLoading(true);
        setError("");
        try {
            const data = await prApi.getPurchaseRequests();
            setPrs(data || []);
            console.log("Purchase requests loaded:", data);
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to load purchase requests";
            console.error("Error loading PRs:", err);
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

    // Filter PRs based on selected status
    const filteredPrs = statusFilter === "ALL" 
        ? prs 
        : prs.filter(pr => pr.Status === statusFilter);

    const statusOptions = ["ALL", "DRAFT", "PENDING", "APPROVED", "REJECTED"];

    return (
        <div className="mx-auto max-w-6xl p-6">
            <div className="mb-6 flex items-center justify-between">
                <h1 className="text-2xl font-bold">Purchase Requests</h1>
                <div className="flex gap-2">
                    <Button onClick={() => setPrFormOpen(true)} className="gap-2">
                        <HugeiconsIcon icon={Add01Icon} strokeWidth={2} className="w-4 h-4" />
                        Create PR
                    </Button>
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
                    <p className="font-medium">Error loading purchase requests:</p>
                    <p className="text-xs mt-1 font-mono">{error}</p>
                </div>
            )}

            {loading && (
                <div className="mb-4 rounded-md bg-blue-50 p-4 text-sm text-blue-700 border border-blue-200">
                    Loading purchase requests...
                </div>
            )}

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>PR Number</TableHead>
                            <TableHead>Department</TableHead>
                            <TableHead>Purpose</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Created At</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {filteredPrs.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center text-muted-foreground">
                                    {prs.length === 0 ? "No purchase requests yet" : `No ${statusFilter === "ALL" ? "" : statusFilter + " "}purchase requests`}
                                </TableCell>
                            </TableRow>
                        ) : (
                            filteredPrs.map((pr) => (
                                <TableRow 
                                    key={pr.ID} 
                                    className={`cursor-pointer hover:bg-muted/50 transition-colors ${pr.DeletedAt ? "opacity-50" : ""}`}
                                    onClick={() => navigate(`/pr/${pr.ID}`)}
                                >
                                    <TableCell className="font-mono font-semibold">
                                        {pr.PRNumber}
                                        {pr.DeletedAt && <span className="ml-2 text-sm text-destructive">(Deleted)</span>}
                                    </TableCell>
                                    <TableCell>{pr.Department}</TableCell>
                                    <TableCell>{pr.Purpose}</TableCell>
                                    <TableCell>
                                        <Badge variant={getStatusColor(pr.Status)}>
                                            {pr.Status}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="text-muted-foreground">
                                        {formatDate(pr.CreatedAt)}
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>

            {/* PR Form Modal */}
            <PrForm
                open={prFormOpen}
                onOpenChange={setPrFormOpen}
                mode="create"
                onSuccess={load}
            />
        </div>
    );
}
    
