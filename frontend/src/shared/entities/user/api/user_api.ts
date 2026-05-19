import {z} from "zod";
import {type addUserFormType, userSchema, type UserType} from "../types/user_types.ts";
import {noteSchema} from "../../notes/types/notes_types.ts";
import {complaintSchema} from "../../complaints/types/complaints_types.ts";
import {EventSchema} from "../../events/types/events_types.ts";
import {apiClient} from "../../../../core/api/client.ts";

const getUserResponseSchema = z.object({
    User: userSchema,
    Notes: z.array(noteSchema).optional(),
    Complaints: z.array(complaintSchema).optional(),
    Events: z.array(EventSchema).optional(),
    ClassTeacher: userSchema.nullable().optional(),
});

export type GettedUser = z.infer<typeof getUserResponseSchema>;

export const UserApi = {
    async getUser(id: number): Promise<GettedUser> {
        const response = await apiClient.get(`/api/users?id=${id}`, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении данных о пользователе");
        }
        
        const parsed = getUserResponseSchema.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный формат пользователя");
        }
        
        return parsed.data
    },
    
    async getStaff(): Promise<UserType[]> {
        const response = await apiClient.get("/api/users/get/staff", true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении персонала");
        }
        
        const parsed = z.array(userSchema).safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }
        
        return parsed.data
    },
    
    async addUser(form: addUserFormType): Promise<boolean> {
      const response = await apiClient.patch("/api/user/add", {
          Name: form.Name,
          LastName: form.LastName,
          FullName: form.FullName,
          Password: form.Password,
          ClassID: form.ClassID,
          Login: form.Login,
          Role: form.Role,
          Rating: 100,
      }, true);  
      
      if (!response.checkStatus()) {
          throw new Error("Ошибка при добавлении пользователя");
      }
      
      return true;
    },
    
    async deleteUser(id: number): Promise<boolean> {
        const response = await apiClient.delete(`/api/user/delete/${id}`, {}, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке удаления пользователя");
        }
        
        return true;
    }
}