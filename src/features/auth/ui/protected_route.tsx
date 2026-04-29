import { Navigate, Outlet } from "react-router-dom";
import { useAuthStore } from "../store/auth_store.ts";

export function ProtectedRoute() {
    const status = useAuthStore((state) => state.status);

    if (status === "idle" || status === "loading") {
        return <div>Loading...</div>;
    }

    if (status !== "authenticated") {
        return <Navigate to="/login" replace />;
    }

    return <Outlet />;
}