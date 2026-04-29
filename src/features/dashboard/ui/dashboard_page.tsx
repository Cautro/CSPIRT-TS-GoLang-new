import { useEffect } from "react";
import { useAuthStore } from "../../auth/store/auth_store.ts";
import { UserDashboardPage } from "./user_dashboard_page.tsx";
import { useDashboardStore } from "../store/dashboard_store.ts";

export function DashboardPage() {
    const role = useAuthStore((state) => state.user?.Role);

    const users = useDashboardStore((state) => state.users) ?? [];
    const notes = useDashboardStore((state) => state.notes) ?? [];
    const status = useDashboardStore((state) => state.status);
    const error = useDashboardStore((state) => state.error);
    const getUsers = useDashboardStore((state) => state.getUsers);
    const getNotes = useDashboardStore((state) => state.getNotes);

    useEffect(() => {
        void getUsers();
        void getNotes();
        console.log(notes);
    }, [getUsers, getNotes]);

    if (!role) {
        return null;
    }

    return (
        <UserDashboardPage
            notes={notes}
            role={role}
            users={users}
            error={error}
            status={status}
            getUsers={getUsers}
        />
    );
}