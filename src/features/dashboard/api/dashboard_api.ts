import { ApiClient } from "../../../core/api/api_client.ts";
import type { UserType } from "../../../shared/entities/user/user_types.ts";

interface ErrorResponse {
    error: string;
}

interface RatingResponse {
    new_rating: number;
    message: string;
}

export interface RatingChangeDTO {
    rating: number;
    target_login: string;
    reason: string;
}

const client = new ApiClient();

function isErrorResponse(data: unknown): data is ErrorResponse {
    return (
        typeof data === "object" &&
        data !== null &&
        "error" in data &&
        typeof (data as ErrorResponse).error === "string"
    );
}

export const dashboardApi = {
    async getUsers(token: string): Promise<UserType[]> {
        const response = await client.get<UserType[] | ErrorResponse>("/api/users", token);

        if (!response.checkStatus()) {
            throw new Error("Ошибка сервера");
        }

        if (isErrorResponse(response.data)) {
            throw new Error(response.data.error);
        }

        if (!Array.isArray(response.data)) {
            throw new Error("Некорректный формат пользователей");
        }

        return response.data;
    },

    async changeRating(token: string, data: RatingChangeDTO): Promise<RatingResponse> {
        const response = await client.patch<RatingResponse | ErrorResponse>(
            "/api/rating/update",
            data,
            token,
        );

        if (!response.checkStatus()) {
            throw new Error("Ошибка сервера");
        }

        if (isErrorResponse(response.data)) {
            throw new Error(response.data.error);
        }

        return response.data as RatingResponse;
    },
};