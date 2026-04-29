import { AppConfig } from "../app_core/app_config.ts";
import { ApiResponse } from "./api_response.ts";

export type ApiMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

interface RequestOptions {
    token?: string | null;
    data?: unknown;
}

export class ApiClient {
    private buildUrl(endpoint: string): string {
        const baseUrl = AppConfig.API_URL.replace(/\/+$/, "");
        const path = endpoint.replace(/^\/+/, "");

        return `${baseUrl}/${path}`;
    }

    private assertSecureUrl(url: string) {
        const target = new URL(url, window.location.origin);
        const isLocalhost =
            target.hostname === "localhost" ||
            target.hostname === "127.0.0.1";

        if (AppConfig.IS_PROD && target.protocol !== "https:" && !isLocalhost) {
            throw new Error("Небезопасное HTTP-соединение запрещено в production");
        }
    }

    private async request<T>(
        method: ApiMethod,
        endpoint: string,
        options: RequestOptions = {},
    ): Promise<ApiResponse<T>> {
        const url = this.buildUrl(endpoint);
        this.assertSecureUrl(url);

        const controller = new AbortController();
        const timeoutId = window.setTimeout(
            () => controller.abort(),
            AppConfig.REQUEST_TIMEOUT_MS,   
        );

        const headers = new Headers({
            Accept: "application/json",
        });

        if (method !== "GET" && options.data !== undefined) {
            headers.set("Content-Type", "application/json");
        }

        if (AppConfig.AUTH_MODE === "bearer-memory" && options.token) {
            headers.set("Authorization", `Bearer ${options.token}`);
        }

        const config: RequestInit = {
            method,
            headers,
            signal: controller.signal,
            cache: "no-store",
            referrerPolicy: "strict-origin-when-cross-origin",
            credentials: AppConfig.AUTH_MODE === "cookie" ? "include" : "same-origin",
        };

        if (method !== "GET" && options.data !== undefined) {
            config.body = JSON.stringify(options.data);
        }

        try {
            const response = await fetch(url, config);
            return ApiResponse.fromResponse<T>(response);
        } finally {
            window.clearTimeout(timeoutId);
        }
    }

    get<T>(endpoint: string, token?: string | null): Promise<ApiResponse<T>> {
        return this.request<T>("GET", endpoint, { token });
    }

    post<T>(
        endpoint: string,
        data: unknown,
        token?: string | null,
    ): Promise<ApiResponse<T>> {
        return this.request<T>("POST", endpoint, { data, token });
    }

    patch<T>(
        endpoint: string,
        data: unknown,
        token?: string | null,
    ): Promise<ApiResponse<T>> {
        return this.request<T>("PATCH", endpoint, { data, token });
    }
}