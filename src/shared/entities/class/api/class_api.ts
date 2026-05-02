import {z} from 'zod'
import {ApiClient} from "../../../../core/api/api_client.ts";
import {classSchema, type ClassType} from "../types/class_types.ts";
import {userSchema, type UserType} from "../../user/types/user_types.ts";

// const errorResponseSchema = z.object({
//     error: z.string().optional(),
//     message: z.string().optional(),
// }).passthrough();

const classesResponseSchema = z.object({
    Classes: z.array(classSchema)
});

const classUsersResponseSchema = z.object({
    Users: z.array(userSchema),
});

const client = new ApiClient();

export const classApi = {
    async getClasses(): Promise<ClassType[]> {
        const response = await client.get("/api/classes", true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении списка классов");
        }
        
        const parsed = classesResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный формат классов");
        }
        
        return parsed.data.Classes;
    },
    
    async getUsersByClass(className: string): Promise<UserType[]> {
        const response = await client.get(`/api/classes/${className}/users`, true);    
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении списка учениокв");
        }
        
        const parsed = classUsersResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный формат пользователей");
        }
        
        return parsed.data.Users;
    }
};