import { useEffect, useState } from "react";
import { vendorApi, type Vendor } from "@/lib/api/vendor";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
    DialogTrigger,
} from "@/components/ui/dialog";
import { toast } from "sonner";

const emptyForm = { name: "", address: "", tax_id: "" };

export default function VendorPage() {
    const [vendors, setVendors] = useState<Vendor[]>([]);
    const [form, setForm] = useState(emptyForm);
    const [editId, setEditId] = useState<number | null>(null);
    const [open, setOpen] = useState(false);
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const load = async () => {
        try {
            const data = await vendorApi.getVendors();
            setVendors(data || []);
        } catch (err) {
            console.error("Failed to load vendors:", err);
            toast.error("Failed to load vendors");
        }
    };

    useEffect(() => {
        load();
    }, []);

    const openCreate = () => {
        setForm(emptyForm);
        setEditId(null);
        setError("");
        setOpen(true);
    };

    const openEdit = (vendor: Vendor) => {
        setForm({ name: vendor.Name, address: vendor.Address, tax_id: vendor.TaxID });
        setEditId(vendor.ID);
        setError("");
        setOpen(true);
    };

    const handleSave = async () => {
        setError("");
        if (!form.name.trim()) {
            setError("Vendor name is required");
            return;
        }
        if (!form.tax_id.trim()) {
            setError("Tax ID is required");
            return;
        }

        setLoading(true);
        try {
            if (editId) {
                await vendorApi.updateVendor(editId, {
                    name: form.name,
                    address: form.address,
                    tax_id: form.tax_id,
                });
                toast.success("Vendor updated successfully");
            } else {
                await vendorApi.createVendor({
                    name: form.name,
                    address: form.address,
                    tax_id: form.tax_id,
                });
                toast.success("Vendor created successfully");
            }
            setOpen(false);
            load();
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to save vendor";
            setError(errorMsg);
            toast.error(errorMsg);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm("Are you sure you want to delete this vendor?")) return;
        try {
            await vendorApi.deleteVendor(id);
            toast.success("Vendor deleted successfully");
            load();
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : "Failed to delete vendor";
            toast.error(errorMsg);
        }
    };

    const set = (key: string, value: string) =>
        setForm((prev) => ({ ...prev, [key]: value }));

    return (
        <div className="mx-auto max-w-5xl p-6">
            <div className="mb-6 flex items-center justify-between">
                <h1 className="text-2xl font-bold">Vendor Management</h1>
                <Dialog open={open} onOpenChange={setOpen}>
                    <DialogTrigger asChild>
                        <Button onClick={openCreate}>Add Vendor</Button>
                    </DialogTrigger>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>{editId ? "Edit Vendor" : "New Vendor"}</DialogTitle>
                        </DialogHeader>
                        <div className="grid gap-4 py-4">
                            <div className="grid gap-2">
                                <Label htmlFor="name">Vendor Name *</Label>
                                <Input 
                                    id="name" 
                                    placeholder="Enter vendor name" 
                                    value={form.name} 
                                    onChange={(e) => set("name", e.target.value)} 
                                />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="address">Address</Label>
                                <Input 
                                    id="address" 
                                    placeholder="Enter vendor address" 
                                    value={form.address} 
                                    onChange={(e) => set("address", e.target.value)} 
                                />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="tax_id">Tax ID *</Label>
                                <Input 
                                    id="tax_id" 
                                    placeholder="Enter tax ID" 
                                    value={form.tax_id} 
                                    onChange={(e) => set("tax_id", e.target.value)} 
                                />
                            </div>
                            {error && <p className="text-sm text-destructive">{error}</p>}
                        </div>
                        <DialogFooter>
                            <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
                            <Button onClick={handleSave} disabled={loading}>
                                {loading ? "Saving..." : editId ? "Update" : "Create"}
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Vendor Name</TableHead>
                            <TableHead>Address</TableHead>
                            <TableHead>Tax ID</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {vendors.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={4} className="text-center text-muted-foreground">
                                    No vendors yet
                                </TableCell>
                            </TableRow>
                        ) : (
                            vendors.map((vendor) => (
                                <TableRow key={vendor.ID}>
                                    <TableCell className="font-semibold">{vendor.Name}</TableCell>
                                    <TableCell className="text-muted-foreground">{vendor.Address || "—"}</TableCell>
                                    <TableCell className="font-mono text-sm">{vendor.TaxID}</TableCell>
                                    <TableCell className="text-right">
                                        <div className="flex justify-end gap-2">
                                            <Button size="sm" variant="outline" onClick={() => openEdit(vendor)}>
                                                Edit
                                            </Button>
                                            <Button size="sm" variant="destructive" onClick={() => handleDelete(vendor.ID)}>
                                                Delete
                                            </Button>
                                        </div>
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
