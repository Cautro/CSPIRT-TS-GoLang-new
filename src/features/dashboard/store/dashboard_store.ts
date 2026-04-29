import { create } from "zustand";
import type { UserType } from "../../../shared/entities/user/user_types.ts";
import { dashboardApi, type RatingChangeDTO } from "../api/dashboard_api.ts";
import { useAuthStore } from "../../auth/store/auth_store.ts";

export type DashboardStatus = "loading" | "error" | "idle";

interface DashboardState {
    status: DashboardStatus;
    users: UserType[];
    error: string | null;
    message: string | null;

    getUsers: () => Promise<void>;
    changeRating: (data: RatingChangeDTO) => Promise<void>;
}

export const useDashboardStore = create<DashboardState>()((set, get) => ({
    error: null,
    status: "idle",
    users: [],
    message: null,

    getUsers: async () => {
        set({ status: "loading", error: null });

        try {
            const token = useAuthStore.getState().token;

            if (!token) {
                throw new Error("Токен авторизации не найден");
            }

            const response = await dashboardApi.getUsers();

            set({
                users: response,
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

    changeRating: async (data: RatingChangeDTO) => {
        set({ status: "loading", error: null, message: null });

        try {
            const token = useAuthStore.getState().token;

            if (!token) {
                throw new Error("Токен авторизации не найден");
            }

            const response = await dashboardApi.changeRating (data);

            set({
                status: "idle",
                error: null,
                message: response.message,
            });

            await get().getUsers();
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
}));