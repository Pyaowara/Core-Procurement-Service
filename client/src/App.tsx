import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider, useAuth } from "@/context/AuthContext";
import Layout from "@/components/Layout";
import LoginPage from "@/pages/LoginPage";
import RegisterPage from "@/pages/RegisterPage";
import InventoryPage from "@/pages/InventoryPage";
import AdminPage from "@/pages/AdminPage";
import ProfilePage from "@/pages/ProfilePage";
import CatalogPage from "@/pages/CatalogPage";
import PrListPage from "@/pages/PrListPage";
import PrDetailPage from "@/pages/PrDetailPage";
import PoListPage from "@/pages/PolistPage";
import PoDetailPage from "@/pages/PoDetailPage";
import VendorPage from "@/pages/VendorPage";
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
            <Route path="/pr" element={<PrListPage />} />
            <Route path="/pr/:id" element={<PrDetailPage />} />
            <Route path="/profile" element={<ProfilePage />} />
            <Route
              path="/po"
              element={
                <RoleGuard roles={["PurchaseOfficer", "Admin"]}>
                  <PoListPage />
                </RoleGuard>
              }
            />
            <Route
              path="/po/:id"
              element={
                <RoleGuard roles={["PurchaseOfficer", "Admin"]}>
                  <PoDetailPage />
                </RoleGuard>
              }
            />
            <Route
              path="/inventory"
              element={
                <RoleGuard roles={["PurchaseOfficer"]}>
                  <InventoryPage />
                </RoleGuard>
              }
            />
            <Route
              path="/vendor"
              element={
                <RoleGuard roles={["Admin", "PurchaseOfficer"]}>
                  <VendorPage />
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
