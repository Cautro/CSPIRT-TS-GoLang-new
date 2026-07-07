import {useMutation, useQueryClient} from "@tanstack/react-query";
import {NotesApi} from "../../../shared/entities/notes/api/notes_api.ts";
import type {NoteFormType} from "../../../shared/entities/notes/types/notes_types.ts";

export const useAddNote = () => {
    const queryclient = useQueryClient();
    
    return useMutation({
        mutationFn: ({form}: {form: NoteFormType}) => NotesApi.addNote(form),
        onSuccess: (_data, variables) => {
            const {form} = variables;
            queryclient.invalidateQueries({queryKey: ["notes"]});
            queryclient.invalidateQueries({queryKey: ["user_by_id", form.TargetID]});
        },
    })
}