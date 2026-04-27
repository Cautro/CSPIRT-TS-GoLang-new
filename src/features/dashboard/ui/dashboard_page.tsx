import {useAuthStore} from "../../auth/store/auth_store.ts";
import {UserDashboardPage} from "./user_dashboard_page.tsx";

export function DashboardPage() {
    const role = useAuthStore(((state) => state.user?.Role));
    
    if (role === null) {
        return null;
    }
    
    switch (role) {
        case "User":
            return <UserDashboardPage/>     
        default:
            return <UserDashboardPage/>
    }
    
}