import {create} from "zustand";
import type {ComplaintType} from "../../../shared/entities/complaints/types/complaints_types.ts";
import {ComplaintsApi} from "../../../shared/entities/complaints/api/complaints_api.ts";

type status = "idle" | "loading" | "error";

interface State {
    status: status;
    error: string | null;
    message: string | null;
    complaints: ComplaintType[] | null;

    getComplaints: (id: number) => Promise<void>;
    deleteComplaint: (id: number) => Promise<void>;
}

export const useComplaintsStore = create<State>()((set) => ({
    status: "idle",
    error: null,
    message: null,
    complaints: null,

    getComplaints: async (id: number) => {
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

    deleteComplaint: async (id: number) => {
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
}));