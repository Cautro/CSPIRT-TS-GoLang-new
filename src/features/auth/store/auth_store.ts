import { create } from "zustand";
import { authApi, type AuthDto } from "../api/auth_api";
import type { UserType } from "../../../shared/entities/user/user_types.ts";
import { AppConfig } from "../../../core/app_core/app_config.ts";

type AuthStatus =
    | "idle"
    | "loading"
    | "authenticated"
    | "unauthenticated";

interface AuthState {
    token: string | null;
    user: UserType | null;
    error: string | null;
    status: AuthStatus;

    login: (dto: AuthDto) => Promise<boolean>;
    checkAuth: () => Promise<void>;
    logout: () => void;
    clearError: () => void;
}

function getPublicErrorMessage(error: unknown): string {
    if (!(error instanceof Error)) return "Ошибка";

    const allowedMessages = new Set([
        "Некорректный логин или пароль",
        "Ошибка авторизации",
        "Сессия недействительна",
        "Некорректный формат профиля",
        "Некорректный ответ сервера авторизации",
    ]);

    return allowedMessages.has(error.message) ? error.message : "Ошибка";
}

export const useAuthStore = create<AuthState>()((set, get) => ({
    token: null,
    user: null,
    error: null,
    status: "idle",

    login: async (dto) => {
        set({
            status: "loading",
            error: null,
        });

        try {
            const token =
                AppConfig.AUTH_MODE === "bearer-memory"
                    ? await authApi.login(dto)
                    : null;


            set({
                token: token,
                status: "authenticated",
                error: null,
            });

            const userData = await authApi.checkAuth();
            
            set({
               user: userData, 
            });
            
            return true;
        } catch (error) {
            set({
                token: null,
                user: null,
                status: "unauthenticated",
                error: getPublicErrorMessage(error),
            });

            return false;
        }
    },

    checkAuth: async () => {
        const token = get().token;

        if (AppConfig.AUTH_MODE === "bearer-memory" && !token) {
            set({
                token: null,
                user: null,
                status: "unauthenticated",
                error: null,
            });

            return;
        }

        set({
            status: "loading",
            error: null,
        });

        try {
            const userData = await authApi.checkAuth();

            set({
                status: "authenticated",
                error: null,
                user: userData,
            });
        } catch {
            set({
                token: null,
                user: null,
                status: "unauthenticated",
                error: "Сессия недействительна",
            });
        }
    },

    logout: () => {
        set({
            token: null,
            user: null,
            status: "unauthenticated",
            error: null,
        });
    },

    clearError: () => {
        set({ error: null });
    },
}));