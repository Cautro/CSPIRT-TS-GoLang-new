import {create} from "zustand"
import type {UserType} from "../../../shared/entities/user/types/user_types.ts";
import type {NoteType} from "../../../shared/entities/notes/types/notes_types.ts";
import {type changeClassTeacherType, classApi} from "../../../shared/entities/class/api/class_api.ts";
import {NotesApi} from "../../../shared/entities/notes/api/notes_api.ts";
import type {ComplaintType} from "../../../shared/entities/complaints/types/complaints_types.ts";
import {ComplaintsApi} from "../../../shared/entities/complaints/api/complaints_api.ts";
import {UserApi} from "../../../shared/entities/user/api/user_api.ts"; 

export type ClassDashboardStatus = "loading" | "error" | "idle";

interface State {
    status: ClassDashboardStatus;
    error: string | null;
    message: string | null;
    users: UserType[];
    notes: NoteType[];
    complaints: ComplaintType[];
    staff: UserType[];
    teacher: UserType | null;
    
    getUsersByClass: (name: string) => Promise<void>
    getNotesByClass: (id: string) => Promise<void>
    deleteNote: (id: string) => Promise<void>
    getComplaints: (id: string) => Promise<void>
    deleteComplaint: (id: string) => Promise<void>
    changeTeacher: (id: number, dto: changeClassTeacherType) => Promise<void>
    getStaff: () => Promise<void>
    getClassTeacher: (id: number) => Promise<void>
}

export const useClassDashboardStore = create<State>()((set) => ({
    status: "idle",
    error: null,
    message: null,
    users: [],
    notes: [],
    complaints: [],
    staff: [],
    teacher: null,

    getUsersByClass: async (name: string) => {
        set({status: "loading"});

        try {
            const response = await classApi.getUsersByClass(name);

            set({status: "idle", users: response, error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },

    getNotesByClass: async (id: string) => {
        set({status: "loading"});

        try {
            const response = await NotesApi.getNotes(id);

            set({status: "idle", notes: response});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },

    deleteNote: async (id: string) => {
        set({status: "loading"});

        try {
            await NotesApi.deleteNote(id);
            set({status: "idle", message: "Заметка успешно удалена", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },

    getComplaints: async (id: string) => {
        set({status: "loading"});
        
        try {
            const response = await ComplaintsApi.getComplaints(id);
            set({status: "idle", error: null, complaints: response});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    deleteComplaint: async (id: string) => {
        set({status: "loading"});

        try {
            await ComplaintsApi.deleteComplaint(id);
            set({status: "idle", message: "Жалоба успешно удалена", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    changeTeacher: async (id: number, dto: changeClassTeacherType) => {
        set({status: "loading"});
        
        try {
            await classApi.changeClassTeacher(id, dto);
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
    }
}))