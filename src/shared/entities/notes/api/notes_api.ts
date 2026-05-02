import {z} from "zod";
import {noteSchema, type NoteType} from "../types/notes_types.ts";
import {ApiClient} from "../../../../core/api/api_client.ts";

export const noteAddDto = z.object({
    AuthorID: z.number().int().nonnegative(),
    CreatedAt: z.string(),
    TargetID: z.number().int().nonnegative(),
    Content: z.string().max(500),
    AuthorName: z.string(),
    TargetName: z.string(),
});

export type noteAddType = z.infer<typeof noteAddDto>

const notesResponseShema = z.object({
    Notes: z.array(noteSchema),
});

const noteAddResponseSchema = z.object({
    message: z.string(),
});

const noteDeleteResponseSchema = z.object({
    message: z.string(),
});

const client = new ApiClient();

export const NotesApi = {
    async getNotes(id: string): Promise<NoteType[]> {
      const response = await client.get(`/api/notes?class=${id}`, true);
      
      if (!response.checkStatus()) {
          throw new Error("Ошибка при получении списка заметок");
      }
      
      const parsed = notesResponseShema.safeParse(response.data);

      if (!parsed.success) {
          throw new Error("Некорректный формат заметок");
      }
      
      return parsed.data.Notes;
    },
    
    async addNote(dto: noteAddType): Promise<boolean> {
        const response = await client.patch("/api/note/add", dto, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при добавлении заметки");
        }
        
        const parsed = noteAddResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }
        
        return true;
    },
    
    async deleteNote(id: string): Promise<boolean> {
        const response = await client.delete(`/api/note/delete/${id}`, {}, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при удалении заметки");
        }

        const parsed = noteDeleteResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }
        
        return true;
    }
}