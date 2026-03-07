import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import Navbar from "@/components/Navbar";

export default function Layout() {
    const { user, loading } = useAuth();

    if (loading) {
        return (
            <div className="flex min-h-screen items-center justify-center">
                <p className="text-muted-foreground">Loading…</p>
            </div>
        );
    }

    if (!user) return <Navigate to="/login" replace />;

    return (
        <div className="min-h-screen">
            <Navbar />
            <main>
                <Outlet />
            </main>
        </div>
    );
}
