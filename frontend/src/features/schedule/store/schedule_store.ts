import {create} from "zustand";
import {
    type ScheduleAddFormValues,
    ScheduleAddLessonUsecase
} from "../models/add_schedule_usecase.ts";
import type {UserType} from "../../../shared/entities/user/types/user_types.ts";
import {
    type ScheduleChangeFormValues,
    ScheduleChangeLessonUsecase
} from "../models/schedule_change_usecase.ts";
import {UserApi} from "../../../shared/entities/user/api/user_api.ts";
import {ScheduleApi} from "../../../shared/entities/schedule/api/schedule_api.ts";
import type {ScheduleModelType} from "../../../shared/entities/schedule/types/schedule_types.ts";
import {ScheduleService} from "../../../shared/entities/schedule/service/schedule_service.ts";

interface State {
    status: "idle" | "error" | "loading",
    error: string | null
    teachers: UserType[] | null
    schedule: ScheduleModelType | null
    
    getClassSchedule: (id: number, type: "base" | "current" | "planned") => Promise<void>
    getTeacherSchedule: (id: number) => Promise<void>
    addSchedule: (form: ScheduleAddFormValues, type: "base" | "current" | "planned") => Promise<boolean>
    changeSchedule: (id: number, form: ScheduleChangeFormValues, type: "base" | "current" | "planned") => Promise<boolean>
    deleteSchedule: (id: number, type: "base" | "current" | "planned") => Promise<void>
    getTeachers: () => Promise<void>
}

export const useScheduleStore = create<State>()((set) => ({
    status: "idle",
    error: null,
    teachers: null,
    schedule: null,

    getClassSchedule: async (id: number, type: "base" | "current" | "planned") => {
        set({status: "loading", schedule: null, error: null});

        try {
            const response = await ScheduleApi.getCurrentScheduleByClass(id, type);
            const schedule = ScheduleService.sortSchedule(response);
            set({status: "idle", error: null, schedule: schedule});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    getTeacherSchedule: async (id: number) => {
        set({status: "loading", schedule: null, error: null});

        try {
            const response = await ScheduleApi.getTeacherSchedule(id);
            const schedule = ScheduleService.sortSchedule(response);
            set({status: "idle", error: null, schedule: schedule});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },

    changeSchedule: async (id: number, form: ScheduleChangeFormValues, type: "base" | "current" | "planned") => {
        set({status: "loading", error: null});

        try {
            const response = await ScheduleChangeLessonUsecase(id, form, type);
            if (response) {
                set({status: "idle", error: null});
                return true;
            }
            return false;
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
            return false
        }
    },

    addSchedule: async (form: ScheduleAddFormValues, type: "base" | "current" | "planned") => {
        set({status: "loading", error: null});

        try {
            const response = await ScheduleAddLessonUsecase(form, type);
            if (response) {
                set({status: "idle", error: null});
                return true;
            }
            return false;
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
            return false
        }
    },

    deleteSchedule: async (id: number, type: "base" | "current" | "planned") => {
        set({status: "loading"});
        try {
            await ScheduleApi.deleteSchedule(id, type);
            set({status: "idle", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },

    getTeachers: async () => {
        set({status: "loading", teachers: null, error: null});

        try {
            const response = await UserApi.getStaff();
            set({error: null, status: "idle", teachers: response});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    }
}))