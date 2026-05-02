import {z} from "zod";
import {ApiClient} from "../../../../core/api/api_client.ts";
import {userSchema} from "../types/user_types.ts";
import {noteSchema} from "../../notes/types/notes_types.ts";
import {complaintSchema} from "../../complaints/types/complaints_types.ts";

const getUserResponseSchema = z.object({
    User: userSchema,
    Notes: z.array(noteSchema).optional(),
    Complaints: z.array(complaintSchema).optional(),
    Events: z.array(noteSchema).optional(),
    ClassTeacher: userSchema.optional(),
})

export type GettedUser = z.infer<typeof getUserResponseSchema>

const client = new ApiClient();

export const UserApi = {
    async getUser(id: string): Promise<GettedUser> {
        
        const response = await client.get(`/api/users/?id=${id}`, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении данных о пользователе");
        }
        
        const parsed = getUserResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный формат пользователя");
        }
        
        return parsed.data
    }
}