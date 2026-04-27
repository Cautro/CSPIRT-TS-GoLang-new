import {AppConfig} from "../app_core/app_config.ts";
import {ApiResponse} from "./api_response.ts";

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

    private async request<T>(
        method: ApiMethod,
        endpoint: string,
        options: RequestOptions = {},
    ): Promise<ApiResponse<T>> {
        const headers = new Headers({
            "Content-Type": "application/json",
        });

        if (options.token) {
            headers.set("Authorization", `Bearer ${options.token}`);
        }

        const config: RequestInit = {
            method,
            headers,
        };

        if (method !== "GET" && options.data !== undefined) {
            config.body = JSON.stringify(options.data);
        }

        const response = await fetch(this.buildUrl(endpoint), config);

        return ApiResponse.fromResponse(response);
    }

    get<T>(endpoint: string, token?: string | null): Promise<ApiResponse<T>> {
        return this.request<T>("GET", endpoint, { token });
    }

    post<T>(
        endpoint: string,
        data: unknown,
        token?: string | null,
    ): Promise<ApiResponse<T>> {
        return this.request<T>("POST", endpoint, {
            data,
            token,
        });
    }
}