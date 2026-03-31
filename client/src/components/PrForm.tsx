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

import { HugeiconsIcon } from "@hugeicons/react"
import { Add01Icon} from "@hugeicons/core-free-icons"
import { userApi, type PublicInventoryItem, prApi, type PurchaseRequest } from "@/lib/api"
import { toast } from "react-hot-toast"
import PrItemCard from "@/components/PrItemCard"

interface PRItemFormData {
    sku: string
    quantity: number
    price_per_unit: number
    discount?: number
    discount_unit?: string
    required_date: string
}

interface FormItem extends PRItemFormData {
    tempId: string // For identifying items before submission
    inventoryItemId?: number
    itemName?: string
    description?: string
    isNewItem?: boolean // Track if this is a new custom item
}

interface PrFormProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    onSuccess?: () => void
    mode: "create" | "edit"
    existingPr?: PurchaseRequest
}

export default function PrForm({
    open,
    onOpenChange,
    onSuccess,
    mode,
    existingPr,
}: PrFormProps) {
    const navigate = useNavigate()
    const [inventoryItems, setInventoryItems] = useState<PublicInventoryItem[]>([])
    const [loadingInventory, setLoadingInventory] = useState(false)
    const [submitting, setSubmitting] = useState(false)
    const [error, setError] = useState("")

    // Form state
    const [department, setDepartment] = useState("")
    const [purpose, setPurpose] = useState("")
    const [items, setItems] = useState<FormItem[]>([])

    useEffect(() => {
        if (open) {
            loadInventoryItems()
            if (mode === "edit" && existingPr) {
                // Pre-fill form with existing PR data
                setDepartment(existingPr.Department)
                setPurpose(existingPr.Purpose)
                const convertedItems: FormItem[] = existingPr.Items.map((item, index) => ({
                    tempId: `existing-${index}`,
                    sku: item.SKU,
                    quantity: item.Quantity,
                    price_per_unit: item.PricePerUnit,
                    discount: item.Discount || 0,
                    discount_unit: item.DiscountUnit || "%",
                    required_date: item.RequiredDate.includes("T") ? item.RequiredDate.split("T")[0] : item.RequiredDate,
                    itemName: item.ItemName,
                    description: item.Description,
                    isNewItem: item.SKU === "", // Mark as new item if SKU is empty
                }))
                setItems(convertedItems)
            } else {
                // Reset for create mode
                setDepartment("")
                setPurpose("")
                setItems([])
            }
            setError("")
        }
    }, [open, mode, existingPr])

    const loadInventoryItems = async () => {
        try {
            setLoadingInventory(true)
            const items = await userApi.getPublicInventory()
            setInventoryItems(items || [])
        } catch (err) {
            console.error("Error loading inventory items:", err)
            toast.error("Failed to load inventory items")
        } finally {
            setLoadingInventory(false)
        }
    }

    const addItem = () => {
        const newItem: FormItem = {
            tempId: `new-${Date.now()}`,
            sku: "",
            quantity: 1,
            price_per_unit: 0,
            discount: 0,
            discount_unit: "%",
            required_date: new Date().toISOString().split("T")[0],
        }
        setItems([...items, newItem])
    }

    const removeItem = (tempId: string) => {
        setItems(items.filter((item) => item.tempId !== tempId))
    }

    const handleSelectInventory = (tempId: string, skuValue: string) => {
        if (skuValue === "new") {
            // New item - allow custom input
            setItems(
                items.map((item) =>
                    item.tempId === tempId
                        ? {
                              ...item,
                              sku: "",
                              itemName: "",
                              description: "",
                              isNewItem: true,
                              inventoryItemId: undefined,
                          }
                        : item
                )
            )
        } else {
            // Existing inventory item
            const selectedInventoryItem = inventoryItems.find(
                (item) => item.sku === skuValue
            )
            if (selectedInventoryItem) {
                setItems(
                    items.map((item) =>
                        item.tempId === tempId
                            ? {
                                  ...item,
                                  inventoryItemId: selectedInventoryItem.id,
                                  sku: selectedInventoryItem.sku,
                                  itemName: selectedInventoryItem.name,
                                  description: selectedInventoryItem.description,
                                  isNewItem: false,
                              }
                            : item
                    )
                )
            }
        }
    }

    const updateItemField = (
        tempId: string,
        field: keyof FormItem,
        value: any
    ) => {
        setItems(
            items.map((item) =>
                item.tempId === tempId
                    ? { ...item, [field]: value }
                    : item
            )
        )
    }

    const calculateTotalPrice = (item: FormItem) => {
        let total = item.quantity * item.price_per_unit
        const discount = item.discount || 0
        if (item.discount_unit === "%") {
            total -= (total * discount) / 100
        } else {
            total -= discount
        }
        return total.toFixed(2)
    }

    const getTodayDate = () => {
        const today = new Date()
        return today.toISOString().split("T")[0]
    }

    const handleSubmit = async () => {
        // Validation
        if (!department.trim()) {
            setError("Department is required")
            return
        }
        if (!purpose.trim()) {
            setError("Purpose is required")
            return
        }
        if (items.length === 0) {
            setError("At least one item is required")
            return
        }

        // Validate all items
        for (const item of items) {
            if (!item.itemName || !item.itemName.trim()) {
                setError("Item name is required for all items")
                return
            }
            if (!item.isNewItem && !item.sku) {
                setError("All items must have an inventory item selected or create a new one")
                return
            }
            if (item.quantity <= 0) {
                setError("Quantity must be greater than 0")
                return
            }
            if (item.price_per_unit < 0) {
                setError("Price per unit cannot be negative")
                return
            }
            if (!item.required_date) {
                setError("Required date is required for all items")
                return
            }
        }

        setSubmitting(true)
        try {
            const prData = {
                department,
                purpose,
                items: items.map((item) => ({
                    sku: item.sku || ``,
                    item_name: item.itemName || "",
                    description: item.description || "",
                    quantity: item.quantity,
                    price_per_unit: item.price_per_unit,
                    discount: item.discount || 0,
                    discount_unit: item.discount_unit || "%",
                    required_date: item.required_date,
                })),
            }

            if (mode === "create") {
                const response = await prApi.createPurchaseRequest(prData)
                toast.success("Purchase Request created successfully")
                onOpenChange(false)
                onSuccess?.()
                // Redirect to PR detail page
                navigate(`/pr/${response.data.ID}`)
            } else {
                // Edit mode - update existing PR
                if (existingPr) {
                    await prApi.editPurchaseRequest(existingPr.ID, prData)
                    toast.success("Purchase Request updated successfully")
                    onOpenChange(false)
                    onSuccess?.()
                }
            }
        } catch (err) {
            const errorMsg =
                err instanceof Error
                    ? err.message
                    : "Failed to submit Purchase Request"
            console.error("Error submitting PR:", err)
            setError(errorMsg)
            toast.error(errorMsg)
        } finally {
            setSubmitting(false)
        }
    }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>
                        {mode === "create"
                            ? "Create Purchase Request"
                            : "Edit Purchase Request"}
                    </DialogTitle>
                </DialogHeader>

                <div className="space-y-6">
                    {/* Department & Purpose */}
                    <div className="grid gap-4 grid-cols-1 md:grid-cols-2">
                        <div className="space-y-2">
                            <Label htmlFor="department">Department *</Label>
                            <Input
                                id="department"
                                placeholder="e.g., Engineering, HR"
                                value={department}
                                onChange={(e) => setDepartment(e.target.value)}
                                disabled={submitting}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="purpose">Purpose *</Label>
                            <Input
                                id="purpose"
                                placeholder="e.g., Office supplies, Equipment"
                                value={purpose}
                                onChange={(e) => setPurpose(e.target.value)}
                                disabled={submitting}
                            />
                        </div>
                    </div>

                    {/* Items Section */}
                    <div className="space-y-3">
                        <div className="flex items-center justify-between">
                            <Label className="text-base font-semibold">
                                Items *
                            </Label>
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={addItem}
                                disabled={submitting}
                                className="gap-2"
                            >
                                <HugeiconsIcon
                                    icon={Add01Icon}
                                    strokeWidth={2}
                                    className="w-4 h-4"
                                />
                                Add Item
                            </Button>
                        </div>

                        {items.length === 0 ? (
                            <div className="rounded-md border border-dashed p-4 text-center text-muted-foreground">
                                <p>
                                    No items added yet. Click "Add Item" to get
                                    started.
                                </p>
                            </div>
                        ) : (
                            <div className="space-y-3">
                                {items.map((item, index) => (
                                    <PrItemCard
                                        key={item.tempId}
                                        item={item}
                                        index={index}
                                        inventoryItems={inventoryItems}
                                        loadingInventory={loadingInventory}
                                        submitting={submitting}
                                        onRemove={removeItem}
                                        onSelectInventory={handleSelectInventory}
                                        onUpdateField={updateItemField}
                                        calculateTotalPrice={calculateTotalPrice}
                                        getTodayDate={getTodayDate}
                                    />
                                ))}
                            </div>
                        )}
                    </div>
                    {error && (
                        <div className="rounded-md bg-destructive/10 p-4 text-sm text-destructive border border-destructive/20">
                            <p className="font-medium">Error:</p>
                            <p className="text-xs mt-1 font-mono">{error}</p>
                        </div>
                    )}                </div>

                <DialogFooter>
                    <Button
                        variant="outline"
                        onClick={() => onOpenChange(false)}
                        disabled={submitting}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleSubmit}
                        disabled={submitting || loadingInventory}
                    >
                        {submitting
                            ? mode === "create"
                                ? "Creating..."
                                : "Updating..."
                            : mode === "create"
                              ? "Create PR"
                              : "Update PR"}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
