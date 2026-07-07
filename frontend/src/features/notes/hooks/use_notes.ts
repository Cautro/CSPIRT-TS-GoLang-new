import {useQuery} from "@tanstack/react-query";
import {NotesApi} from "../../../shared/entities/notes/api/notes_api.ts";

export const useNotes = (id: number)=> {
    return useQuery({
        queryKey: ["notes", id],
        queryFn: () => NotesApi.getNotes(id),
    });
}