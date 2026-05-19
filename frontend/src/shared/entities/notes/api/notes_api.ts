import {z} from "zod";
import {noteSchema, type NoteType} from "../types/notes_types.ts";
import {apiClient} from "../../../../core/api/client.ts";

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

export const NotesApi = {
    async getNotes(id: number): Promise<NoteType[]> {
      const response = await apiClient.get(`/api/notes?class=${id}`, true);
      
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
        const response = await apiClient.patch("/api/note/add", dto, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при добавлении заметки");
        }
        
        return true;
    },
    
    async deleteNote(id: number): Promise<boolean> {
        const response = await apiClient.delete(`/api/note/delete/${id}`, {}, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при удалении заметки");
        }

        return true;
    }
}