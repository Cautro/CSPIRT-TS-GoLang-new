import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { authApi, type AuthDto } from "../api/auth_api";
import type {UserType} from "../../../shared/entities/user/user_types.ts";

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

function getErrorMessage(error: unknown): string {
    if (error instanceof Error) return error.message;
    return "Неизвестная ошибка";
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set, get) => ({
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
                    const token = await authApi.login(dto);

                    set({
                        token,
                        status: "authenticated",
                        error: null,
                    });

                    return true;
                } catch (error) {
                    set({
                        token: null,
                        status: "unauthenticated",
                        error: getErrorMessage(error),
                    });

                    return false;
                }
            },

            checkAuth: async () => {
                const token = get().token;

                if (!token) {
                    set({
                        token: null,
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
                    const userData = await authApi.checkAuth(token);

                    if (userData) {
                        set({
                            status: "authenticated",
                            error: null,
                            user: userData,
                        });
                    } else {
                        set({
                            token: null,
                            status: "unauthenticated",
                            error: null,
                        });
                    }
                } catch {
                    set({
                        token: null,
                        status: "unauthenticated",
                        error: "Сессия недействительна",
                    });
                }
            },

            logout: () => {
                set({
                    token: null,
                    status: "unauthenticated",
                    error: null,
                });
            },

            clearError: () => {
                set({ error: null });
            },
        }),
        {
            name: "auth-storage",
            storage: createJSONStorage(() => localStorage),

            partialize: (state) => ({
                token: state.token,
                user: state.user,
            }),
        },
    ),
);