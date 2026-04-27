import { ApiClient } from "../../../core/api/api_client";
import type {UserType} from "../../../shared/entities/user/user_types.ts";

export interface AuthDto {
    login: string;
    password: string;
}

interface LoginResponse {
    token: string;
}

interface ErrorResponse {
    error?: string;
    message?: string;
}

const client = new ApiClient();

export const authApi = {
    async login(dto: AuthDto): Promise<string> {
        const response = await client.post<LoginResponse | ErrorResponse>(
            "/login",
            {
                Login: dto.login,
                Password: dto.password,
            },
        );

        const json = response.data;

        if (!response.checkStatus()) {
            const message =
                json && "error" in json && json.error
                    ? json.error
                    : "Ошибка авторизации";

            throw new Error(message);
        }

        if (!json || !("token" in json) || !json.token) {
            throw new Error("Сервер не вернул токен");
        }

        return json.token;
    },

    async checkAuth(token: string): Promise<UserType> {
        const response = await client.get<unknown>("/api/me", token);
        if (!response.checkStatus()) {
            throw new Error("Ошибка авторизации")
        }
        return response.data as UserType;
    },
};