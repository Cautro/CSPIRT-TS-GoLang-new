import {z} from 'zod'
import {type addClassFormType, classSchema, type ClassType} from "../types/class_types.ts";
import {userSchema, type UserType} from "../../user/types/user_types.ts";
import {apiClient} from "../../../../core/api/client.ts";

const classesResponseSchema = z.object({
    Classes: z.array(classSchema)
});

const classUsersResponseSchema = z.object({
    Users: z.array(userSchema),
});

export const classTeacherResponseSchema = z.object({
    Teacher: userSchema
});

export const classApi = {
    async getClasses(): Promise<ClassType[]> {
        const response = await apiClient.get("/api/classes", true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении списка классов");
        }
        
        const parsed = classesResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный формат классов");
        }
        
        return parsed.data.Classes.sort((a, b) => (b.UserTotalRating + b.ClassTotalRating) - (a.ClassTotalRating + a.UserTotalRating));
    },
    
    async getClassById(id: number): Promise<ClassType> {
        const response = await apiClient.get(`/api/classes?class_id=${id}`, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении класса");
        }

        const parsed = classesResponseSchema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный формат классов");
        }
        
        if (parsed.data.Classes.length === 0) {
            throw new Error("Класс не найден");
        }

        return parsed.data.Classes[0];
    },
    
    async getUsersByClass(id: number): Promise<UserType[]> {
        const response = await apiClient.get(`/api/classes/${id}/users`, true);    
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении списка учениокв");
        }
        
        const parsed = classUsersResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный формат пользователей");
        }
        
        return parsed.data.Users;
    },
    
    async changeClassTeacher(id: number, teacher: string): Promise<boolean> {
        const response = await apiClient.patch(`/api/classes/${id}/teacher`, {
            TeacherLogin: teacher
        }, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при изменении классного руководителя");
        }
        
        return true;
    },
    
    async getClassTeacher(classId: number): Promise<UserType> {
        const response = await apiClient.get(`/api/classes/${classId}/teacher`, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении классного руководителя");
        }
        
        const parsed = classTeacherResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }
        return parsed.data.Teacher;
    },
    
    async addClass(dto: addClassFormType): Promise<boolean> {
        const response = await apiClient.patch(`/api/classes/add`, dto, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при добавлении класса");
        }
        
        return true;
    },
    
    async deleteClass(id: number): Promise<boolean> {
        const response = await apiClient.delete(`/api/classes/delete/${id}`, {}, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке удаления класса");
        }
        
        return true;
    }
};