import { useEffect, useState } from "react";
import { userApi, type InventoryItem } from "@/lib/api/index";
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

const emptyForm = { sku: "", name: "", description: "", quantity: 0 };

export default function InventoryPage() {
    const [items, setItems] = useState<InventoryItem[]>([]);
    const [form, setForm] = useState(emptyForm);
    const [editId, setEditId] = useState<number | null>(null);
    const [open, setOpen] = useState(false);
    const [error, setError] = useState("");

    const load = async () => {
        const data = await userApi.getInventory();
        setItems(data || []);
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

    const openEdit = (item: InventoryItem) => {
        setForm({ sku: item.Sku, name: item.Name, description: item.Description, quantity: item.Quantity });
        setEditId(item.ID);
        setError("");
        setOpen(true);
    };

    const handleSave = async () => {
        setError("");
        try {
            if (editId) {
                await userApi.updateInventory(editId, form);
            } else {
                await userApi.createInventory(form);
            }
            setOpen(false);
            load();
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : "Failed to save");
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm("Delete this item?")) return;
        await userApi.deleteInventory(id);
        load();
    };

    const set = (key: string, value: string | number) =>
        setForm((prev) => ({ ...prev, [key]: value }));

    return (
        <div className="mx-auto max-w-4xl p-6">
            <div className="mb-6 flex items-center justify-between">
                <h1 className="text-2xl font-bold">Inventory Management</h1>
                <Dialog open={open} onOpenChange={setOpen}>
                    <DialogTrigger asChild>
                        <Button onClick={openCreate}>Add Item</Button>
                    </DialogTrigger>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>{editId ? "Edit Item" : "New Item"}</DialogTitle>
                        </DialogHeader>
                        <div className="grid gap-4 py-4">
                            <div className="grid gap-2">
                                <Label htmlFor="sku">SKU</Label>
                                <Input id="sku" value={form.sku} onChange={(e) => set("sku", e.target.value)} />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="name">Name</Label>
                                <Input id="name" value={form.name} onChange={(e) => set("name", e.target.value)} />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="description">Description</Label>
                                <Input id="description" value={form.description} onChange={(e) => set("description", e.target.value)} />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="quantity">Quantity</Label>
                                <Input id="quantity" type="number" value={form.quantity} onChange={(e) => set("quantity", Number(e.target.value))} />
                            </div>
                            {error && <p className="text-sm text-destructive">{error}</p>}
                        </div>
                        <DialogFooter>
                            <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
                            <Button onClick={handleSave}>{editId ? "Update" : "Create"}</Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>SKU</TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead>Description</TableHead>
                            <TableHead className="text-right">Qty</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {items.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center text-muted-foreground">
                                    No inventory items yet
                                </TableCell>
                            </TableRow>
                        ) : (
                            items.map((item) => (
                                <TableRow key={item.ID}>
                                    <TableCell className="font-mono">{item.Sku}</TableCell>
                                    <TableCell>{item.Name}</TableCell>
                                    <TableCell className="text-muted-foreground">{item.Description}</TableCell>
                                    <TableCell className="text-right">{item.Quantity}</TableCell>
                                    <TableCell className="text-right">
                                        <div className="flex justify-end gap-2">
                                            <Button size="sm" variant="outline" onClick={() => openEdit(item)}>Edit</Button>
                                            <Button size="sm" variant="destructive" onClick={() => handleDelete(item.ID)}>Delete</Button>
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
