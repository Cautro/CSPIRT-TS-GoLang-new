import {useAuthStore} from "../store/auth_store.ts";
import {Navigate, Outlet} from "react-router-dom";

export function ProtectedRoute() {
    const token = useAuthStore((state) => state.token);
    const status = useAuthStore((state) => state.status);
    
    if (status === "idle" || status === "loading") {
        return <div>Loading...</div>
    }
    
    if (!token) {
        return <Navigate to="/login" replace />;
    } 
    
    return <Outlet/>
}