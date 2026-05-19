import {create} from "zustand";
import type {UserType} from "../../../shared/entities/user/types/user_types.ts";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";
import {UserApi} from "../../../shared/entities/user/api/user_api.ts";
import {AddUserUseCase, type addUserValues} from "../models/add_user_usecase.ts";

type status = "idle" | "loading" | "error";

interface State {
    status: status;
    error: string | null;
    message: string | null;
    users: UserType[] | null;
    staff: UserType[] | null;

    getUsersByClass: (id: number) => Promise<void>;    
    getStaff: () => Promise<void>;
    addUser: (form: addUserValues) => Promise<boolean>
}

export const useUsersStore = create<State>()((set) => ({
    status: "idle",
    error: null,
    message: null,
    users: null,
    staff: null,   
    
    getUsersByClass: async (id: number)=> {
        set({status: "loading", users: null, error: null});

        try {
            const response = await classApi.getUsersByClass(id);

            set({status: "idle", users: response, error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },

    getStaff: async () => {
        set({status: "loading", staff: null, error: null});

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

    addUser: async (form: addUserValues) => {
        set({status: "loading", error: null});

        try {
            const response = await AddUserUseCase(form);
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
    }
    
}));