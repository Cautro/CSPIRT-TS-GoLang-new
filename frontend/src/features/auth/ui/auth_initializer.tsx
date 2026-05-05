import { useEffect, type ReactNode } from "react";
import {useAuthStore} from "../store/auth_store.ts";

interface AuthInitializerProps {
    children: ReactNode;
}

export function AuthInitializer({ children }: AuthInitializerProps) {
    const refreshAuth = useAuthStore((state) => state.refreshAuth);

    useEffect(() => {
        void refreshAuth();
    }, [refreshAuth]);

    return <>{children}</>;
}