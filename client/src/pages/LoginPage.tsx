import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import InteractiveGridBackground from "@/components/lightswind/interactive-grid-background";

function redirectForRole() {
    return "/dashboard";
}

export default function LoginPage() {
    const navigate = useNavigate();
    const { login } = useAuth();
    const [form, setForm] = useState({ username: "", password: "" });
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        setLoading(true);
        try {
            await login(form);
            navigate(redirectForRole());
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : "Login failed");
        } finally {
            setLoading(false);
        }
    };

    return (
        <InteractiveGridBackground>
            <div className="flex min-h-screen w-full items-center justify-center p-4">
                <Card className="z-10 w-full max-w-sm shadow-sm bg-card/80 backdrop-blur-md">
                    <CardHeader className="text-center">
                        <div className="mx-auto mb-2 flex h-12 w-full items-center justify-center text-primary text-lg font-bold">
                            Core Procurement Service
                        </div>
                        <CardDescription>Sign in to your account</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="grid gap-4">
                            <div className="grid gap-2">
                                <Label htmlFor="username">Username</Label>
                                <Input
                                    id="username"
                                    placeholder="Enter your username"
                                    value={form.username}
                                    onChange={(e) => setForm({ ...form, username: e.target.value })}
                                    required
                                />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="password">Password</Label>
                                <Input
                                    id="password"
                                    type="password"
                                    placeholder="Enter your password"
                                    value={form.password}
                                    onChange={(e) => setForm({ ...form, password: e.target.value })}
                                    required
                                />
                            </div>
                            {error && <p className="text-sm text-destructive">{error}</p>}
                            <Button type="submit" className="w-full" disabled={loading}>
                                {loading ? "Signing in…" : "Sign In"}
                            </Button>
                            <p className="text-center text-sm text-muted-foreground">
                                Don&apos;t have an account?{" "}
                                <Link to="/register" className="font-medium text-primary underline underline-offset-4 hover:text-primary/80">
                                    Register
                                </Link>
                            </p>
                        </form>
                    </CardContent>
                </Card>
            </div>
        </InteractiveGridBackground>
    );
}
