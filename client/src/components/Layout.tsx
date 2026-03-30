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
        <div className="flex min-h-screen flex-col bg-background">
            <Navbar />
            <main className="flex-1 pb-8">
                <Outlet />
            </main>
            <footer className="mt-auto bg-black py-4 text-center text-sm text-white/80 rounded-t-lg">
                <p>&copy; {new Date().getFullYear()} Core Procurement Service. All rights reserved.</p>
            </footer>
        </div>
    );
}
