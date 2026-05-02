import {create} from "zustand"
import {type GettedUser, UserApi} from "../../../shared/entities/user/api/user_api.ts";
import {type noteAddType, NotesApi} from "../../../shared/entities/notes/api/notes_api.ts";
import {type complaintAddType, ComplaintsApi} from "../../../shared/entities/complaints/api/complaints_api.ts";
import {RatingApi, type ratingChangeType} from "../../../shared/entities/rating/api/rating_api.ts";

export type UserStatus = "loading" | "error" | "idle";

interface State {
    status: UserStatus;
    error: string | null;
    user: GettedUser | null;
    message: string | null;
    
    getUser: (id: string) => Promise<void>;
    addNote: (dto: noteAddType) => Promise<void>;
    deleteNote: (id: string) => Promise<void>;
    addComplaint: (dto: complaintAddType) => Promise<void>;
    deleteComplaint: (id: string) => Promise<void>;
    changeRating: (dto: ratingChangeType) => Promise<void>;
}

export const useUserStore = create<State>()((set) => ({
    status: "idle",
    error: null,
    user: null,
    message: null,
    
    getUser: async (id: string) => {
        set({status: "idle"});
        
        try {
            const response = await UserApi.getUser(id);
            set({status: "idle", user: response});
            
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    addNote: async (dto: noteAddType) => {
        set({status: "loading"});
        
        try {
            await NotesApi.addNote(dto);
            set({status: "idle", message: "Заметка успешно добавлена", error: null});
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

    addComplaint: async (dto: complaintAddType) => {
        set({status: "loading"});

        try {
            await ComplaintsApi.addComplaint(dto)
            set({status: "idle", message: "Жалоба успешно добавлена", error: null});
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
    
    changeRating: async (dto: ratingChangeType) => {
        set({status: "loading"});
        
        try {
            const response = await RatingApi.changeRating(dto);
            set({status: "idle", message: response, error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    }
}))