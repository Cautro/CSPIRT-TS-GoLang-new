import { AppConfig } from "../app_core/app_config.ts";
import { ApiResponse } from "./api_response.ts";
import {clearAccessToken, getAccessToken, setAccessToken} from "../auth/access_token_memory";
export type ApiMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

interface RequestOptions {
    data?: unknown;
    skipAuthRetry?: boolean;
    auth?: boolean;
    credentials?: RequestCredentials;
}

export class ApiClient {
    private buildUrl(endpoint: string): string {
        const baseUrl = AppConfig.API_URL.replace(/\/+$/, "");
        const path = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;

        if (!baseUrl) {
            return path;
        }

        return `${baseUrl}${path}`;
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

        if (options.auth) {
            const token = getAccessToken();

            if (token) {
                headers.set("Authorization", `Bearer ${token}`);
            }
        }

        const config: RequestInit = {
            method,
            headers,
            signal: controller.signal,
            cache: "no-store",
            referrerPolicy: "strict-origin-when-cross-origin",
            credentials: options.credentials ?? "include",
        };

        if (method !== "GET" && options.data !== undefined) {
            config.body = JSON.stringify(options.data);
        }

        try {
            let response = await fetch(url, config);

            if (
                response.status === 401 &&
                options.auth &&
                !options.skipAuthRetry
            ) {
                const refreshed = await this.refreshAccessToken();

                if (refreshed) {
                    response = await fetch(url, config);
                }
            }
            
            return ApiResponse.fromResponse<T>(response);
        } finally {
            window.clearTimeout(timeoutId);
        }
    }
    
    private async refreshAccessToken(): Promise<boolean> {
        const response = await this.request<{token: string}>(
            "POST",
            "/api/refresh",
            {
                auth: false,
                skipAuthRetry: true,
            }
        );

        if (!response.checkStatus() || !response.data?.token) {
            clearAccessToken();
            return false;
        }

        setAccessToken(response.data.token);
        return true;
    }

    get<T>(endpoint: string, auth = false): Promise<ApiResponse<T>> {
        return this.request<T>("GET", endpoint, { auth });
    }

    post<T>(
        endpoint: string,
        data?: unknown,
        auth = false,
    ): Promise<ApiResponse<T>> {
        return this.request<T>("POST", endpoint, { data, auth });
    }

    patch<T>(
        endpoint: string,
        data?: unknown,
        auth = false,
    ): Promise<ApiResponse<T>> {
        return this.request<T>("PATCH", endpoint, { data, auth });
    }
    
    delete<T>(
        endpoint: string,
        data?: unknown,
        auth = false,
    ): Promise<ApiResponse<T>> {
        return this.request<T>("DELETE", endpoint, { data, auth });
    }
}