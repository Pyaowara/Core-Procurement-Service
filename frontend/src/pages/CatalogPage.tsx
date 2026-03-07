import { useEffect, useState } from "react";
import { api, type PublicInventoryItem } from "@/lib/api";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";

export default function CatalogPage() {
    const [items, setItems] = useState<PublicInventoryItem[]>([]);

    useEffect(() => {
        api.getPublicInventory().then((data) => setItems(data || []));
    }, []);

    return (
        <div className="mx-auto max-w-4xl p-6">
            <h1 className="mb-6 text-2xl font-bold">Catalog</h1>

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>SKU</TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead>Description</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {items.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={3} className="text-center text-muted-foreground">
                                    No items available
                                </TableCell>
                            </TableRow>
                        ) : (
                            items.map((item) => (
                                <TableRow key={item.id}>
                                    <TableCell className="font-mono">{item.sku}</TableCell>
                                    <TableCell>{item.name}</TableCell>
                                    <TableCell className="text-muted-foreground">{item.description}</TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}
