import { useEffect, useState } from "react";
import { userApi, type User } from "@/lib/api/index";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

const ROLES = ["Admin", "PurchaseOfficer", "Manager", "Executive", "Employee"];

export default function AdminPage() {
    const { user: me } = useAuth();
    const [users, setUsers] = useState<User[]>([]);

    const load = async () => {
        const data = await userApi.getUsers();
        setUsers(data || []);
    };

    useEffect(() => {
        load();
    }, []);

    const handleRoleChange = async (userId: number, role: string) => {
        await userApi.updateUserRole(userId, role);
        load();
    };

    const handleDelete = async (id: number) => {
        if (!confirm("Delete this user?")) return;
        await userApi.deleteUser(id);
        load();
    };

    return (
        <div className="mx-auto max-w-4xl p-6">
            <h1 className="mb-6 text-2xl font-bold">User Management</h1>

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>ID</TableHead>
                            <TableHead>Username</TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead>Email</TableHead>
                            <TableHead>Role</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {users.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center text-muted-foreground">
                                    No users found
                                </TableCell>
                            </TableRow>
                        ) : (
                            users.map((u) => (
                                <TableRow key={u.id}>
                                    <TableCell>{u.id}</TableCell>
                                    <TableCell className="font-medium">{u.username}</TableCell>
                                    <TableCell>{u.first_name} {u.last_name}</TableCell>
                                    <TableCell className="text-muted-foreground">{u.email}</TableCell>
                                    <TableCell>
                                        <Select
                                            value={u.role}
                                            onValueChange={(role) => handleRoleChange(u.id, role)}
                                        >
                                            <SelectTrigger className="h-8 w-40">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {ROLES.map((r) => (
                                                    <SelectItem key={r} value={r}>{r}</SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                    </TableCell>
                                    <TableCell className="text-right">
                                        {me && u.id !== me.id && (
                                            <Button size="sm" variant="destructive" onClick={() => handleDelete(u.id)}>
                                                Delete
                                            </Button>
                                        )}
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
