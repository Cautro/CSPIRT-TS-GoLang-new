import {create} from "zustand"
import type {UserType} from "../../../shared/entities/user/types/user_types.ts";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";
import {UserApi} from "../../../shared/entities/user/api/user_api.ts";
import {ScheduleApi} from "../../../shared/entities/schedule/api/schedule_api.ts";
export type ClassDashboardStatus = "loading" | "error" | "idle";

interface State {
    status: ClassDashboardStatus;
    error: string | null;
    message: string | null;
    staff: UserType[];
    teacher: UserType | null;

    changeTeacher: (id: number, teacher: string) => Promise<void>
    getStaff: () => Promise<void>
    getClassTeacher: (id: number) => Promise<void>
    deleteClass: (id: number) => Promise<void>
    rolloverSchedule: (id: number) => Promise<void>
}

export const useClassDashboardStore = create<State>()((set) => ({
    status: "idle",
    error: null,
    message: null,
    staff: [],
    teacher: null,
    
    changeTeacher: async (id: number, teacher: string) => {
        set({status: "loading"});
        
        try {
            await classApi.changeClassTeacher(id, teacher);
            set({status: "idle", message: "Классный руководитель успешно изменён", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    getStaff: async () => {
        set({status: "loading"});

        try {
            const response = await UserApi.getStaff();
            set({staff: response, error: null, status: "idle"});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    getClassTeacher: async (id: number)=> {
        set({status: "loading"});
        
        try {
            const response = await classApi.getClassTeacher(id);
            set({teacher: response, error: null, status: "idle"});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    deleteClass: async (id: number) => {
        set({status: "loading"});
        
        try {
            await classApi.deleteClass(id);
            set({status: "idle", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    
    rolloverSchedule: async (id: number) => {
        set({status: "loading"});
        try {
            await ScheduleApi.rolloverSchedule(id);
            set({status: "idle", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    }
}))