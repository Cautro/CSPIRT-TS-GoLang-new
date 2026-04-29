import { z } from "zod";
import { ApiClient } from "../../../core/api/api_client";
import { userSchema, type UserType } from "../../../shared/entities/user/user_types.ts";
import { LOGIN_REGEX, SECURITY_LIMITS } from "../../../core/security/security_limits.ts";

export interface AuthDto {
    login: string;
    password: string;
}

const authDtoSchema = z.object({
    login: z
        .string()
        .min(SECURITY_LIMITS.loginMin)
        .max(SECURITY_LIMITS.loginMax)
        .regex(LOGIN_REGEX),
    password: z
        .string()
        .min(SECURITY_LIMITS.passwordMin)
        .max(SECURITY_LIMITS.passwordMax),
});

const loginResponseSchema = z.object({
    accessToken: z.string().min(20).max(4096),
});

const refreshResponseSchema = z.object({
    token: z.string().min(20).max(4096),
});

const errorResponseSchema = z.object({
    error: z.string().optional(),
    message: z.string().optional(),
}).passthrough();

const client = new ApiClient();

function getSafeAuthError(data: unknown): string {
    const parsed = errorResponseSchema.safeParse(data);

    if (!parsed.success) {
        return "Ошибка авторизации";
    }

    return parsed.data.error || parsed.data.message || "Ошибка авторизации";
}

export const authApi = {
    async login(dto: AuthDto): Promise<string> {
        const parsedDto = authDtoSchema.safeParse(dto);

        if (!parsedDto.success) {
            throw new Error("Некорректный логин или пароль");
        }

        const response = await client.post<unknown>("/login", {
            Login: parsedDto.data.login,
            Password: parsedDto.data.password,
        });

        if (!response.checkStatus()) {
            throw new Error(getSafeAuthError(response.data));
        }

        const parsed = loginResponseSchema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера авторизации");
        }

        return parsed.data.accessToken;
    },
    
    async refresh(): Promise<string> {
        const response = await client.post<unknown>("api/refresh");
        
        if (!response.checkStatus()) {
            throw new Error("Не удалось обновить сессию");
        }

        const parsed = refreshResponseSchema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный ответ refresh");
        }

        return parsed.data.token;
    },

    async checkAuth(): Promise<UserType> {
        const response = await client.get<unknown>("/api/me", true);

        if (!response.checkStatus()) {
            throw new Error("Сессия недействительна");
        }

        const parsedUser = userSchema.safeParse(response.data);

        if (!parsedUser.success) {
            throw new Error("Некорректный формат профиля");
        }

        return parsedUser.data;
    },
};