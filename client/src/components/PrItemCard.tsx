"use client"

import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { HugeiconsIcon } from "@hugeicons/react";
import { Delete02Icon } from "@hugeicons/core-free-icons";
import { type PublicInventoryItem } from "@/lib/api/user";

interface FormItem {
    tempId: string;
    sku: string;
    quantity: number;
    price_per_unit: number;
    discount?: number;
    discount_unit?: string;
    required_date: string;
    inventoryItemId?: number;
    itemName?: string;
    description?: string;
    isNewItem?: boolean;
}

interface PrItemCardProps {
    item: FormItem;
    index: number;
    inventoryItems: PublicInventoryItem[];
    loadingInventory: boolean;
    submitting: boolean;
    onRemove: (tempId: string) => void;
    onSelectInventory: (tempId: string, skuValue: string) => void;
    onUpdateField: (tempId: string, field: keyof FormItem, value: any) => void;
    calculateTotalPrice: (item: FormItem) => string;
    getTodayDate: () => string;
}

export default function PrItemCard({
    item,
    index,
    inventoryItems,
    loadingInventory,
    submitting,
    onRemove,
    onSelectInventory,
    onUpdateField,
    calculateTotalPrice,
    getTodayDate,
}: PrItemCardProps) {
    return (
        <Card>
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <CardTitle className="text-sm">
                        Item {index + 1}
                    </CardTitle>
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => onRemove(item.tempId)}
                        disabled={submitting}
                        className="text-destructive hover:text-destructive"
                    >
                        <HugeiconsIcon
                            icon={Delete02Icon}
                            strokeWidth={2}
                        />
                    </Button>
                </div>
            </CardHeader>
            <CardContent className="space-y-4">
                {/* SKU Selection */}
                <div className="space-y-2">
                    <Label htmlFor={`sku-${item.tempId}`}>
                        SKU *
                    </Label>
                    <Select
                        value={
                            item.isNewItem
                                ? "new"
                                : item.sku || ""
                        }
                        onValueChange={(value) =>
                            onSelectInventory(
                                item.tempId,
                                value
                            )
                        }
                        disabled={
                            submitting ||
                            loadingInventory
                        }
                    >
                        <SelectTrigger
                            id={`sku-${item.tempId}`}
                        >
                            <SelectValue placeholder="Select an item or create new..." />
                        </SelectTrigger>
                        <SelectContent>
                            {loadingInventory ? (
                                <SelectItem
                                    value="_loading"
                                    disabled
                                >
                                    Loading...
                                </SelectItem>
                            ) : (
                                <>
                                    {inventoryItems.length >
                                        0 && (
                                        inventoryItems.map(
                                            (
                                                invItem
                                            ) => (
                                                <SelectItem
                                                    key={
                                                        invItem.id
                                                    }
                                                    value={
                                                        invItem.sku
                                                    }
                                                >
                                                    <span className="font-semibold">
                                                        {
                                                            invItem.sku
                                                        }{" "}
                                                        -{" "}
                                                        {
                                                            invItem.name
                                                        }
                                                    </span>
                                                </SelectItem>
                                            )
                                        )
                                    )}
                                    <SelectItem value="new">
                                        <span className="font-semibold text-blue-600">
                                            + Create New
                                            Item
                                        </span>
                                    </SelectItem>
                                </>
                            )}
                        </SelectContent>
                    </Select>
                </div>

                {/* Item Name */}
                <div className="space-y-2">
                    <Label
                        htmlFor={`item-name-${item.tempId}`}
                    >
                        Item Name *
                    </Label>
                    <Input
                        id={`item-name-${item.tempId}`}
                        placeholder={
                            item.isNewItem
                                ? "Enter item name"
                                : "Auto-filled from selection"
                        }
                        value={item.itemName || ""}
                        onChange={(e) =>
                            onUpdateField(
                                item.tempId,
                                "itemName",
                                e.target.value
                            )
                        }
                        disabled={
                            submitting ||
                            !item.isNewItem
                        }
                    />
                </div>

                {/* Description */}
                <div className="space-y-2">
                    <Label
                        htmlFor={`description-${item.tempId}`}
                    >
                        Description
                    </Label>
                    <Textarea
                        id={`description-${item.tempId}`}
                        placeholder={
                            item.isNewItem
                                ? "Enter item description"
                                : "Auto-filled from selection"
                        }
                        value={item.description || ""}
                        onChange={(e) =>
                            onUpdateField(
                                item.tempId,
                                "description",
                                e.target.value
                            )
                        }
                        disabled={
                            submitting ||
                            !item.isNewItem
                        }
                    />
                </div>

                {/* Quantity, Price, Required Date */}
                <div className="grid gap-4 grid-cols-1 md:grid-cols-3">
                    <div className="space-y-2">
                        <Label
                            htmlFor={`quantity-${item.tempId}`}
                        >
                            Quantity *
                        </Label>
                        <Input
                            id={`quantity-${item.tempId}`}
                            type="number"
                            min="1"
                            value={item.quantity}
                            onChange={(e) =>
                                onUpdateField(
                                    item.tempId,
                                    "quantity",
                                    parseInt(
                                        e.target
                                            .value
                                    ) || 1
                                )
                            }
                            disabled={submitting}
                        />
                    </div>
                    <div className="space-y-2">
                        <Label
                            htmlFor={`price-${item.tempId}`}
                        >
                            Price Per Unit *
                        </Label>
                        <Input
                            id={`price-${item.tempId}`}
                            type="number"
                            step="0.01"
                            min="0"
                            value={item.price_per_unit}
                            onChange={(e) =>
                                onUpdateField(
                                    item.tempId,
                                    "price_per_unit",
                                    parseFloat(
                                        e.target.value
                                    ) || 0
                                )
                            }
                            disabled={submitting}
                        />
                    </div>
                    <div className="space-y-2">
                        <Label
                            htmlFor={`required-date-${item.tempId}`}
                        >
                            Required Date *
                        </Label>
                        <Input
                            id={`required-date-${item.tempId}`}
                            type="date"
                            value={item.required_date}
                            onChange={(e) =>
                                onUpdateField(
                                    item.tempId,
                                    "required_date",
                                    e.target.value
                                )
                            }
                            min={getTodayDate()}
                            disabled={submitting}
                        />
                    </div>
                </div>

                {/* Discount */}
                <div className="grid gap-4 grid-cols-2">
                    <div className="space-y-2">
                        <Label
                            htmlFor={`discount-${item.tempId}`}
                        >
                            Discount
                        </Label>
                        <Input
                            id={`discount-${item.tempId}`}
                            type="number"
                            step="0.01"
                            min="0"
                            value={item.discount}
                            onChange={(e) =>
                                onUpdateField(
                                    item.tempId,
                                    "discount",
                                    parseFloat(
                                        e.target.value
                                    ) || 0
                                )
                            }
                            disabled={submitting}
                        />
                    </div>
                    <div className="space-y-2">
                        <Label
                            htmlFor={`discount-unit-${item.tempId}`}
                        >
                            Unit
                        </Label>
                        <Select
                            value={item.discount_unit}
                            onValueChange={(value) =>
                                onUpdateField(
                                    item.tempId,
                                    "discount_unit",
                                    value
                                )
                            }
                            disabled={
                                submitting ||
                                item.discount === 0
                            }
                        >
                            <SelectTrigger
                                id={`discount-unit-${item.tempId}`}
                                size="sm"
                            >
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="%">
                                    Percentage (%)
                                </SelectItem>
                                <SelectItem value="BAHT">
                                    Baht
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>

                {/* Total Price */}
                <div className="rounded-md bg-primary/5 p-3 flex items-center justify-between">
                    <span className="text-sm font-medium">
                        Total Price:
                    </span>
                    <span className="font-semibold text-lg">
                        {calculateTotalPrice(item)}
                    </span>
                </div>
            </CardContent>
        </Card>
    );
}
