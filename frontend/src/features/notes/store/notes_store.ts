import {create} from "zustand";
import type {NoteType} from "../../../shared/entities/notes/types/notes_types.ts";
import {NotesApi} from "../../../shared/entities/notes/api/notes_api.ts";

type status = "idle" | "loading" | "error";

interface State {
    status: status;
    error: string | null;
    message: string | null;
    notes: NoteType[] | null;

    getNotes: (id: number) => Promise<void>;
    deleteNote: (id: number) => Promise<void>;
}

export const useNotesStore = create<State>()((set) => ({
    status: "idle",
    error: null,
    message: null,
    notes: null,

    getNotes: async (id: number)=> {
        set({status: "loading"});

        try {
            const response = await NotesApi.getNotes(id);

            set({status: "idle", notes: response, error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },

    deleteNote: async (id: number) => {
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
}));