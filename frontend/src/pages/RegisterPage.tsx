import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { api } from "@/lib/api";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

export default function RegisterPage() {
    const navigate = useNavigate();
    const [form, setForm] = useState({
        username: "",
        password: "",
        first_name: "",
        last_name: "",
        email: "",
    });
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const set = (key: string, value: string) =>
        setForm((prev) => ({ ...prev, [key]: value }));

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        setLoading(true);
        try {
            await api.register(form);
            navigate("/login");
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : "Registration failed");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex min-h-screen items-center justify-center bg-muted/30">
            <Card className="w-full max-w-sm shadow-sm">
                <CardHeader className="text-center">
                    <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-primary text-primary-foreground text-lg font-bold">
                        P
                    </div>
                    <CardTitle className="text-2xl">Create account</CardTitle>
                    <CardDescription>Get started with Procurement</CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="grid gap-4">
                        <div className="grid gap-2">
                            <Label htmlFor="username">Username</Label>
                            <Input id="username" placeholder="Choose a username" value={form.username} onChange={(e) => set("username", e.target.value)} required />
                        </div>
                        <div className="grid grid-cols-2 gap-3">
                            <div className="grid gap-2">
                                <Label htmlFor="first_name">First Name</Label>
                                <Input id="first_name" placeholder="John" value={form.first_name} onChange={(e) => set("first_name", e.target.value)} required />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="last_name">Last Name</Label>
                                <Input id="last_name" placeholder="Doe" value={form.last_name} onChange={(e) => set("last_name", e.target.value)} required />
                            </div>
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="email">Email</Label>
                            <Input id="email" type="email" placeholder="john@example.com" value={form.email} onChange={(e) => set("email", e.target.value)} required />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="password">Password</Label>
                            <Input id="password" type="password" placeholder="Create a password" value={form.password} onChange={(e) => set("password", e.target.value)} required />
                        </div>
                        {error && <p className="text-sm text-destructive">{error}</p>}
                        <Button type="submit" className="w-full" disabled={loading}>
                            {loading ? "Creating account…" : "Create Account"}
                        </Button>
                        <p className="text-center text-sm text-muted-foreground">
                            Already have an account?{" "}
                            <Link to="/login" className="font-medium text-primary underline underline-offset-4 hover:text-primary/80">
                                Sign in
                            </Link>
                        </p>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}
