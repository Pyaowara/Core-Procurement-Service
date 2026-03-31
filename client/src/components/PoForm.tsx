"use client"

import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
import { poApi } from "@/lib/api/po"
import { vendorApi, type Vendor } from "@/lib/api/vendor"
import { prApi, type PurchaseRequest } from "@/lib/api/pr"
import { toast } from "react-hot-toast"

interface PoFormProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    onSuccess?: () => void
    prId: number
}

interface SelectedItem {
    prItemId: number
    selected: boolean
    quantity: number
    price_per_unit: number
    discount: number
    discount_unit: string
    sku?: string // For custom SKU when PR item SKU is empty
    itemName: string // Editable item name
    description?: string // Editable description
}

export default function PoForm({
    open,
    onOpenChange,
    onSuccess,
    prId,
}: PoFormProps) {
    const navigate = useNavigate()
    const [pr, setPr] = useState<PurchaseRequest | null>(null)
    const [vendors, setVendors] = useState<Vendor[]>([])
    const [loadingPr, setLoadingPr] = useState(false)
    const [loadingVendors, setLoadingVendors] = useState(false)
    const [submitting, setSubmitting] = useState(false)
    const [error, setError] = useState("")

    // Form state
    const [selectedVendor, setSelectedVendor] = useState<number | null>(null)
    const [creditDays, setCreditDays] = useState("0")
    const [dueDate, setDueDate] = useState("")
    const [selectedItems, setSelectedItems] = useState<Map<number, SelectedItem>>(new Map())

    useEffect(() => {
        if (open) {
            loadPrAndVendors()
        }
    }, [open, prId])

    const loadPrAndVendors = async () => {
        try {
            setLoadingPr(true)
            setLoadingVendors(true)
            const prData = await prApi.getPurchaseRequestById(prId)
            setPr(prData)
            
            // Initialize selected items with PR item data
            const newSelectedItems = new Map<number, SelectedItem>()
            prData.Items.forEach((item) => {
                newSelectedItems.set(item.ID, {
                    prItemId: item.ID,
                    selected: false,
                    quantity: item.Quantity,
                    price_per_unit: item.PricePerUnit,
                    discount: item.Discount || 0,
                    discount_unit: item.DiscountUnit || "%",
                    sku: item.SKU || "", // Initialize with PR item SKU or empty string
                    itemName: item.ItemName,
                    description: item.Description || "",
                })
            })
            setSelectedItems(newSelectedItems)

            const vendorsData = await vendorApi.getVendors()
            setVendors(vendorsData || [])
            setError("")
        } catch (err) {
            const errorMsg = err instanceof Error ? err.message : "Failed to load data"
            setError(errorMsg)
            toast.error(errorMsg)
        } finally {
            setLoadingPr(false)
            setLoadingVendors(false)
        }
    }

    const handleSelectItem = (itemId: number, checked: boolean) => {
        const updated = new Map(selectedItems)
        const item = updated.get(itemId)
        if (item) {
            item.selected = checked
            updated.set(itemId, item)
            setSelectedItems(updated)
        }
    }

    const handleUpdateField = (itemId: number, field: keyof SelectedItem, value: any) => {
        const updated = new Map(selectedItems)
        const item = updated.get(itemId)
        if (item) {
            (item[field] as any) = value
            updated.set(itemId, item)
            setSelectedItems(updated)
        }
    }

    const calculateTotal = (item: SelectedItem) => {
        const subtotal = item.quantity * item.price_per_unit
        if (item.discount_unit === "%") {
            return subtotal * (1 - item.discount / 100)
        }
        return subtotal - item.discount
    }

    const handleSubmit = async () => {
        if (!selectedVendor) {
            setError("Please select a vendor")
            return
        }

        if (!dueDate) {
            setError("Please set a due date")
            return
        }

        // Validate due date is not in the past
        const today = new Date().toISOString().split("T")[0]
        if (dueDate < today) {
            setError("Due date cannot be in the past")
            return
        }

        // Validate credit days
        const creditDaysNum = parseInt(creditDays)
        if (isNaN(creditDaysNum) || creditDaysNum < 0) {
            setError("Credit days must be 0 or greater")
            return
        }

        const selectedItemsList = Array.from(selectedItems.values())
            .filter(item => item.selected)

        if (selectedItemsList.length === 0) {
            setError("Please select at least one item")
            return
        }

        // Validate that items with empty PR SKU have a SKU entered
        const missingSkus = selectedItemsList.some(item => {
            const prItem = pr?.Items.find(i => i.ID === item.prItemId)
            return !prItem?.SKU && !item.sku
        })

        if (missingSkus) {
            setError("Please enter SKU for all items")
            return
        }

        // Validate quantities, prices, and discounts
        const invalidItem = selectedItemsList.find(item => {
            // Quantity must be > 0
            if (item.quantity <= 0) return true
            // Price per unit must be >= 0
            if (item.price_per_unit < 0) return true
            // Discount must be >= 0
            if (item.discount < 0) return true
            // Discount cannot exceed 100% if using percentage
            if (item.discount_unit === "%" && item.discount > 100) return true
            // Discount amount cannot exceed subtotal if using BAHT
            if (item.discount_unit === "BAHT" && item.discount > item.quantity * item.price_per_unit) return true
            return false
        })

        if (invalidItem) {
            setError("Please check: Qty must be > 0, prices and discounts must be valid")
            return
        }

        setSubmitting(true)
        try {
            const response = await poApi.createPurchaseOrder({
                pr_id: prId,
                vendor_id: selectedVendor,
                po_items: selectedItemsList.map(item => ({
                    pr_item_id: item.prItemId,
                    sku: item.sku || undefined, // Include custom SKU if provided
                    item_name: item.itemName,
                    description: item.description,
                    quantity: item.quantity,
                    price_per_unit: item.price_per_unit,
                    discount: item.discount,
                    discount_unit: item.discount_unit,
                })),
                credit_day: parseInt(creditDays) || 0,
                due_date: dueDate,
            })
            toast.success("Purchase Order created successfully")
            onOpenChange(false)
            onSuccess?.()
            // Redirect to PO detail page
            navigate(`/po/${response.data.ID}`)
        } catch (err: unknown) {
            const errorMsg = err instanceof Error ? err.message : "Failed to create PO"
            setError(errorMsg)
            toast.error(errorMsg)
        } finally {
            setSubmitting(false)
        }
    }

    const formatCurrency = (amount: number) => {
        return new Intl.NumberFormat("th-TH", {
            style: "currency",
            currency: "THB",
        }).format(amount)
    }

    const getTodayDate = () => {
        return new Date().toISOString().split("T")[0]
    }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="!max-w-full !w-[70vw] max-h-[85vh] overflow-y-auto p-4">
                <DialogHeader className="pb-2">
                    <DialogTitle className="text-lg">Generate Purchase Order from PR</DialogTitle>
                </DialogHeader>

                {error && (
                    <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive border border-destructive/20">
                        {error}
                    </div>
                )}

                {loadingPr || loadingVendors ? (
                    <div className="py-8 text-center text-muted-foreground">
                        Loading...
                    </div>
                ) : pr ? (
                    <div className="space-y-4">
                        {/* Vendor Selection */}
                        <div className="grid grid-cols-1 gap-3">
                            <div className="space-y-2">
                                <Label htmlFor="vendor">Vendor *</Label>
                                <Select
                                    value={selectedVendor?.toString() || ""}
                                    onValueChange={(value) => setSelectedVendor(parseInt(value))}
                                    disabled={submitting}
                                >
                                    <SelectTrigger id="vendor">
                                        <SelectValue placeholder="Select a vendor" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {vendors.map((vendor) => (
                                            <SelectItem key={vendor.ID} value={vendor.ID.toString()}>
                                                {vendor.Name}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="grid grid-cols-2 gap-3">
                                <div className="space-y-2">
                                    <Label htmlFor="credit-days">Credit Days</Label>
                                    <Input
                                        id="credit-days"
                                        type="number"
                                        min="0"
                                        value={creditDays}
                                        onChange={(e) => setCreditDays(e.target.value)}
                                        disabled={submitting}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="due-date">Due Date *</Label>
                                    <Input
                                        id="due-date"
                                        type="date"
                                        min={getTodayDate()}
                                        value={dueDate}
                                        onChange={(e) => setDueDate(e.target.value)}
                                        disabled={submitting}
                                    />
                                </div>
                            </div>
                        </div>

                        {/* PR Items Selection */}
                        <div className="space-y-2">
                            <Label>PR Items</Label>
                            <div className="rounded-md border overflow-x-auto">
                                <Table className="text-sm w-full">
                                    <TableHeader>
                                        <TableRow className="bg-gray-50">
                                            <TableHead className="w-10 py-2 text-center">Select</TableHead>
                                            <TableHead className="w-16 py-2 text-center">SKU</TableHead>
                                            <TableHead className="w-24 py-2 text-center">Item Name</TableHead>
                                            <TableHead className="w-32 py-2 text-center">Description</TableHead>
                                            <TableHead className="text-center w-14 py-2">Qty</TableHead>
                                            <TableHead className="text-center w-16 py-2">Price</TableHead>
                                            <TableHead className="text-center w-24 py-2">Discount</TableHead>
                                            <TableHead className="text-center w-16 py-2">Total</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {pr.Items.map((prItem) => {
                                            const selectedItem = selectedItems.get(prItem.ID)
                                            if (!selectedItem) return null

                                            return (
                                                <TableRow key={prItem.ID}>
                                                    <TableCell>
                                                        <Checkbox
                                                            checked={selectedItem.selected}
                                                            onCheckedChange={(checked: boolean) =>
                                                                handleSelectItem(prItem.ID, checked as boolean)
                                                            }
                                                            disabled={submitting}
                                                        />
                                                    </TableCell>
                                                    <TableCell className="font-mono text-sm">
                                                        {prItem.SKU ? (
                                                            <span>{prItem.SKU}</span>
                                                        ) : (
                                                            <Input
                                                                type="text"
                                                                placeholder="Enter SKU"
                                                                value={selectedItem.sku || ""}
                                                                onChange={(e) =>
                                                                    handleUpdateField(prItem.ID, "sku", e.target.value)
                                                                }
                                                                disabled={!selectedItem.selected || submitting}
                                                                className="w-24 text-sm h-9 px-2"
                                                            />
                                                        )}
                                                    </TableCell>
                                                    <TableCell>
                                                        <Input
                                                            type="text"
                                                            placeholder="Item Name"
                                                            value={selectedItem.itemName}
                                                            onChange={(e) =>
                                                                handleUpdateField(prItem.ID, "itemName", e.target.value)
                                                            }
                                                            disabled={!selectedItem.selected || submitting}
                                                            className="w-24 text-sm h-9 px-2"
                                                        />
                                                    </TableCell>
                                                    <TableCell>
                                                        <Input
                                                            type="text"
                                                            placeholder="Description"
                                                            value={selectedItem.description || ""}
                                                            onChange={(e) =>
                                                                handleUpdateField(prItem.ID, "description", e.target.value)
                                                            }
                                                            disabled={!selectedItem.selected || submitting}
                                                            className="w-32 text-sm h-9 px-2"
                                                        />
                                                    </TableCell>
                                                    <TableCell className="text-right">
                                                        <Input
                                                            type="number"
                                                            min="1"
                                                            value={selectedItem.quantity}
                                                            onChange={(e) =>
                                                                handleUpdateField(prItem.ID, "quantity", parseInt(e.target.value) || 1)
                                                            }
                                                            disabled={!selectedItem.selected || submitting}
                                                            className="w-28 text-right text-sm h-9 px-2"
                                                        />
                                                    </TableCell>
                                                    <TableCell className="text-right">
                                                        <Input
                                                            type="number"
                                                            step="0.01"
                                                            value={selectedItem.price_per_unit}
                                                            onChange={(e) =>
                                                                handleUpdateField(prItem.ID, "price_per_unit", parseFloat(e.target.value) || 0)
                                                            }
                                                            disabled={!selectedItem.selected || submitting}
                                                            className="w-32 text-right text-sm h-9 px-2"
                                                        />
                                                    </TableCell>
                                                    <TableCell className="text-right">
                                                        <div className="flex gap-1 justify-end">
                                                            <Input
                                                                type="number"
                                                                step="0.01"
                                                                value={selectedItem.discount}
                                                                onChange={(e) =>
                                                                    handleUpdateField(prItem.ID, "discount", parseFloat(e.target.value) || 0)
                                                                }
                                                                disabled={!selectedItem.selected || submitting}
                                                                className="w-20 text-right text-sm h-9 px-2"
                                                                placeholder="0"
                                                            />
                                                            <Select
                                                                value={selectedItem.discount_unit}
                                                                onValueChange={(value) =>
                                                                    handleUpdateField(prItem.ID, "discount_unit", value)
                                                                }
                                                                disabled={!selectedItem.selected || submitting}
                                                            >
                                                                <SelectTrigger className="w-16 text-sm h-9 px-2">
                                                                    <SelectValue />
                                                                </SelectTrigger>
                                                                <SelectContent className="text-sm">
                                                                    <SelectItem value="%">%</SelectItem>
                                                                    <SelectItem value="BAHT">฿</SelectItem>
                                                                </SelectContent>
                                                            </Select>
                                                        </div>
                                                    </TableCell>
                                                    <TableCell className="text-right font-semibold text-sm">
                                                        {selectedItem.selected
                                                            ? formatCurrency(calculateTotal(selectedItem))
                                                            : "—"}
                                                    </TableCell>
                                                </TableRow>
                                            )
                                        })}
                                    </TableBody>
                                </Table>
                            </div>
                        </div>

                        {/* Summary */}
                        <div className="bg-muted/50 p-3 rounded-lg text-xs space-y-1">
                            <div className="flex justify-between">
                                <span className="text-muted-foreground">Selected Items:</span>
                                <span className="font-semibold">
                                    {Array.from(selectedItems.values()).filter(i => i.selected).length}
                                </span>
                            </div>
                            <div className="flex justify-between border-t pt-1">
                                <span className="text-muted-foreground">Total Amount:</span>
                                <span className="font-bold">
                                    {formatCurrency(
                                        Array.from(selectedItems.values())
                                            .filter(i => i.selected)
                                            .reduce((sum, item) => sum + calculateTotal(item), 0)
                                    )}
                                </span>
                            </div>
                        </div>
                    </div>
                ) : (
                    <div className="py-8 text-center text-destructive">
                        Failed to load PR data
                    </div>
                )}

                <DialogFooter>
                    <Button
                        variant="outline"
                        onClick={() => onOpenChange(false)}
                        disabled={submitting}
                    >
                        Cancel
                    </Button>
                    <Button onClick={handleSubmit} disabled={submitting || loadingPr || loadingVendors}>
                        {submitting ? "Creating..." : "Create PO"}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
