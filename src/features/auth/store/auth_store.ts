import { create } from "zustand";
import { authApi, type AuthDto } from "../api/auth_api";
import type { UserType } from "../../../shared/entities/user/user_types.ts";
import {clearAccessToken, setAccessToken} from "../../../core/auth/access_token_memory.ts";

type AuthStatus =
    | "idle"
    | "loading"
    | "authenticated"
    | "unauthenticated";

interface AuthState {
    user: UserType | null;
    error: string | null;
    status: AuthStatus;

    login: (dto: AuthDto) => Promise<boolean>;
    checkAuth: () => Promise<void>;
    logout: () => void;
    clearError: () => void;
    refreshAuth: () => Promise<void>;
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

export const useAuthStore = create<AuthState>()((set) => ({
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
            
            setAccessToken(token);
            
            const user = await authApi.checkAuth();

            set({
                status: "authenticated",
                error: null,
                user: user,
            });
            
            return true;
        } catch (error) {
            set({
                user: null,
                status: "unauthenticated",
                error: getPublicErrorMessage(error),
            });

            return false;
        }
    },

    checkAuth: async () => {
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
            clearAccessToken();
            
            set({
                user: null,
                status: "unauthenticated",
                error: "Сессия недействительна",
            });
        }
    },
    
    refreshAuth: async () => {
      set({status: "loading", error: null});
      
      try {
          const token = await authApi.refresh();
          
          setAccessToken(token);
          
          const user = await authApi.checkAuth();
          
          set({
             status: "authenticated",
             error: null,
             user: user, 
          });
          
      } catch {
          clearAccessToken();

          set({
              user: null,
              status: "unauthenticated",
              error: null,
          });
      }
    },

    logout: () => {
        clearAccessToken();
        
        set({
            user: null,
            status: "unauthenticated",
            error: null,
        });
    },

    clearError: () => {
        set({ error: null });
    },
}));