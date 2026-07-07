import type {UserType} from "../../../shared/entities/user/types/user_types.ts";
import {noteAddFormSchema} from "./note_add_schema.ts";
import type {NoteFormType} from "../../../shared/entities/notes/types/notes_types.ts";

interface NoteAddFormValues {
    current_user: UserType,
    user: UserType,
    content: string,
}


export const NoteAddUsecase = (form: NoteAddFormValues): NoteFormType => {
    const dto = {
        TargetID: form.user.Id,
        Content: form.content.trim(),
        AuthorID: form.current_user.Id,
        CreatedAt: new Date().toISOString(),
        AuthorName: `${form.current_user.Name} ${form.current_user.LastName}`,
        TargetName: `${form.user.Name} ${form.user.LastName}`,
    };

    const parsed = noteAddFormSchema.safeParse(dto);

    if (!parsed.success) {
        throw new Error("Проверьте текст заметки");
    }
    
    return parsed.data;
}