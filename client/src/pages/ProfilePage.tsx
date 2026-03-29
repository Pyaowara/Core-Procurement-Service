import { useEffect, useState } from "react";
import { authApi, userApi, type User } from "@/lib/api/index";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
} from "@/components/ui/card";


export default function ProfilePage() {
    const { user, refresh } = useAuth();
    const [fullUser, setFullUser] = useState<User | null>(null);
    const [form, setForm] = useState({ first_name: "", last_name: "", email: "" });
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState("");
    const [error, setError] = useState("");

    useEffect(() => {
        if (!user) return;
        authApi.me().then((data) => {
            console.log(user,user.user_id)
            setFullUser(data);
            setForm({ first_name: data.first_name, last_name: data.last_name, email: data.email });
            setLoading(false);
        });
    }, [user]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;
        setError("");
        setMessage("");
        setSaving(true);
        try {
            await userApi.updateUser(user.user_id, form);
            setMessage("Profile updated successfully");
            refresh();
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : "Failed to update");
        } finally {
            setSaving(false);
        }
    };

    const set = (key: string, value: string) =>
        setForm((prev) => ({ ...prev, [key]: value }));

    if (loading) {
        return (
            <div className="flex min-h-[50vh] items-center justify-center">
                <p className="text-muted-foreground">Loading…</p>
            </div>
        );
    }

    return (
        <div className="mx-auto max-w-md p-6 pt-10">
            <Card className="shadow-sm">
                <CardHeader className="text-center">
                    <div className="mx-auto mb-2 flex h-16 w-16 items-center justify-center rounded-full bg-primary/10 text-2xl font-bold text-primary">
                        {fullUser?.first_name?.[0]}{fullUser?.last_name?.[0]}
                    </div>
                    <CardDescription className="flex items-center justify-center gap-2">
                        <Badge variant="outline">{user?.role}</Badge>
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="grid gap-4">
                        <div className="grid grid-cols-2 gap-3">
                            <div className="grid gap-2">
                                <Label htmlFor="first_name">First Name</Label>
                                <Input id="first_name" value={form.first_name} onChange={(e) => set("first_name", e.target.value)} required />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="last_name">Last Name</Label>
                                <Input id="last_name" value={form.last_name} onChange={(e) => set("last_name", e.target.value)} required />
                            </div>
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="email">Email</Label>
                            <Input id="email" type="email" value={form.email} onChange={(e) => set("email", e.target.value)} required />
                        </div>
                        {error && <p className="text-sm text-destructive">{error}</p>}
                        {message && <p className="text-sm text-green-600">{message}</p>}
                        <Button type="submit" className="w-full" disabled={saving}>
                            {saving ? "Saving…" : "Save Changes"}
                        </Button>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}
