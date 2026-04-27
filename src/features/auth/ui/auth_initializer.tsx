import { useEffect, type ReactNode } from "react";
import {useAuthStore} from "../store/auth_store.ts";

interface AuthInitializerProps {
    children: ReactNode;
}

export function AuthInitializer({ children }: AuthInitializerProps) {
    const checkAuth = useAuthStore((state) => state.checkAuth);

    useEffect(() => {
        void checkAuth();
    }, [checkAuth]);

    return <>{children}</>;
}