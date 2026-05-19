import {
    type ScheduleAddLessonFormType,
    type ScheduleChangeLessonFormType,
    type ScheduleLessonType,
    SchedulesResponseSchema, TeacherScheduleResponseSchema
} from "../types/schedule_types.ts";
import {apiClient} from "../../../../core/api/client.ts";

export const ScheduleApi = {
    async getCurrentScheduleByClass(id: number, type: "base" | "current" | "planned"): Promise<ScheduleLessonType[]> {
        const response = await apiClient.get(`/api/schedules?class_id=${id}&type=${type}`, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении расписания");
        } 
        
        const parsed = SchedulesResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        } 
        
        return parsed.data.Schedules;
    },
    
    async getTeacherSchedule(id: number): Promise<ScheduleLessonType[]> {
        const response = await apiClient.get(`/api/schedules/teacher/current?teacher_id=${id}`, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении расписания");
        }

        const parsed = TeacherScheduleResponseSchema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }

        return parsed.data.Schedules;
    },
    
    async changeScheduleLesson(id: number, form: ScheduleChangeLessonFormType, type: "base" | "current" | "planned"): Promise<true> {
        const response = await apiClient.patch(`/api/schedules/update`, {
            Type: type,
            Action: "upsert",
            Id: id,
            Lesson: form
        }, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке изменении расписания");
        }

        return true;
    },
    
    async addScheduleLesson(form: ScheduleAddLessonFormType, type: "base" | "current" | "planned"): Promise<true> {
        const response = await apiClient.patch(`/api/schedules/update`, {
            Type: type,
            Action: "upsert",
            Lesson: form
        }, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке изменении расписания");
        }

        return true;
    },
    
    async deleteSchedule(id: number, type: "base" | "current" | "planned"): Promise<true> {
        const response = await apiClient.patch(`/api/schedules/update`, {
            Type: type,
            Action: "delete",
            Id: id 
        }, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке изменении расписания");
        }

        return true;
    },
    
    async rolloverSchedule(id: number): Promise<true> {
        const response = await apiClient.patch(`/api/schedules/rollover?class_id=${id}`, {}, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке сброса текущего расписания");
        }

        return true;
    },
}