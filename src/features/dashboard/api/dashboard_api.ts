import {ApiClient} from "../../../core/api/api_client.ts";
import type {UserType} from "../../../shared/entities/user/user_types.ts";

interface ErrorResponse {
    error: string;
}

const client = new ApiClient();

export const dashboardApi = {
    async getUsers (token: string): Promise<UserType[] | ErrorResponse> {
        const response = await client.get("/api/users", token);
        console.log(response.data);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка сервера");
        }
        
        if (!(response.data as UserType[]) || (response.data as UserType[]).length === 0) {
            throw new Error("Не удалось получить список пользователей");
        }
        
        const json = response.data as UserType[];
        
        return json;
    }
}