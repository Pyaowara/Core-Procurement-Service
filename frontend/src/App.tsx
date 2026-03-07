import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider, useAuth } from "@/context/AuthContext";
import Layout from "@/components/Layout";
import LoginPage from "@/pages/LoginPage";
import RegisterPage from "@/pages/RegisterPage";
import InventoryPage from "@/pages/InventoryPage";
import AdminPage from "@/pages/AdminPage";
import ProfilePage from "@/pages/ProfilePage";
import CatalogPage from "@/pages/CatalogPage";
import RoleGuard from "@/components/RoleGuard";

function HomeRedirect() {
  const { user, loading } = useAuth();
  if (loading) return null;
  if (!user) return <Navigate to="/login" replace />;
  if (user.role === "Admin") return <Navigate to="/admin" replace />;
  if (user.role === "PurchaseOfficer") return <Navigate to="/inventory" replace />;
  return <Navigate to="/catalog" replace />;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          <Route element={<Layout />}>
            <Route path="/catalog" element={<CatalogPage />} />
            <Route path="/profile" element={<ProfilePage />} />
            <Route
              path="/inventory"
              element={
                <RoleGuard roles={["PurchaseOfficer"]}>
                  <InventoryPage />
                </RoleGuard>
              }
            />
            <Route
              path="/admin"
              element={
                <RoleGuard roles={["Admin"]}>
                  <AdminPage />
                </RoleGuard>
              }
            />
          </Route>

          <Route path="*" element={<HomeRedirect />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
