import {z} from 'zod'
import {ApiClient} from "../../../../core/api/api_client.ts";
import {classSchema, type ClassType} from "../types/class_types.ts";
import {userSchema, type UserType} from "../../user/types/user_types.ts";

const classesResponseSchema = z.object({
    Classes: z.array(classSchema)
});

const classUsersResponseSchema = z.object({
    Users: z.array(userSchema),
});

export const changeClassTeacherDto = z.object({
    TeacherLogin: z.string()
});

export type changeClassTeacherType = z.infer<typeof changeClassTeacherDto>

export const classTeacherResponseSchema = z.object({
    Teacher: userSchema
});

export const addClassDto = z.object({
    Name: z.string(),
    TeacherLogin: z.string(),
});

export type addClassType = z.infer<typeof addClassDto>

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
    
    async getClassById(id: number): Promise<ClassType> {
        const response = await client.get(`/api/classes/?class_id=${id}`, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении класса");
        }

        const parsed = classesResponseSchema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный формат классов");
        }

        return parsed.data.Classes[0];
    },
    
    async getUsersByClass(id: number): Promise<UserType[]> {
        const response = await client.get(`/api/classes/${id}/users`, true);    
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении списка учениокв");
        }
        
        const parsed = classUsersResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный формат пользователей");
        }
        
        return parsed.data.Users;
    },
    
    async changeClassTeacher(id: number, dto: changeClassTeacherType): Promise<boolean> {
        const response = await client.patch(`/api/classes/${id}/teacher`, dto, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при изменении классного руководителя");
        }
        
        return true;
    },
    
    async getClassTeacher(classId: number): Promise<UserType> {
        const response = await client.get(`/api/classes/${classId}/teacher`, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении классного руководителя");
        }
        
        const parsed = classTeacherResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }
        return parsed.data.Teacher;
    },
    
    async addClass(dto: addClassType): Promise<boolean> {
        const response = await client.patch(`/api/classes/add`, dto, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при добавлении класса");
        }
        
        return true;
    }
};