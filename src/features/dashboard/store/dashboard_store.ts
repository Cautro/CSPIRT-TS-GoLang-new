import { create } from "zustand";
import type { UserType } from "../../../shared/entities/user/user_types.ts";
import { dashboardApi } from "../api/dashboard_api.ts";

type DashboardStatus = "loading" | "error" | "idle";

interface DashboardState {
    status: DashboardStatus;
    users: UserType[];
    error: string | null;
    getUsers: () => Promise<void>;
}

export const useDashboardStore = create<DashboardState>()((set) => ({
    error: null,
    status: "idle",
    users: [],

    getUsers: async () => {
        set({ status: "loading", error: null });

        try {
            const authDataRaw = localStorage.getItem("auth-storage");
            if (!authDataRaw) {
                throw new Error("Данные авторизации отсутствуют");
            }
            const authData = JSON.parse(authDataRaw);
            const token = authData?.state?.token || authData?.token;

            if (!token) {
                throw new Error("Токен авторизации не найден");
            }

            const response = await dashboardApi.getUsers(token);

            set({
                users: response as UserType[],
                status: "idle",
                error: null,
            });
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
}));
