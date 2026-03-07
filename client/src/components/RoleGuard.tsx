import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";

interface Props {
    roles: string[];
    children: ReactNode;
}

export default function RoleGuard({ roles, children }: Props) {
    const { user } = useAuth();

    if (!user) return <Navigate to="/login" replace />;
    if (!roles.includes(user.role)) return <Navigate to="/catalog" replace />;

    return <>{children}</>;
}
