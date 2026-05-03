import {z} from "zod";
import {ApiClient} from "../../../../core/api/api_client.ts";
import {fullNameSchema, userSchema, type UserType} from "../types/user_types.ts";
import {noteSchema} from "../../notes/types/notes_types.ts";
import {complaintSchema} from "../../complaints/types/complaints_types.ts";

const getUserResponseSchema = z.object({
    User: userSchema,
    Notes: z.array(noteSchema).optional(),
    Complaints: z.array(complaintSchema).optional(),
    Events: z.array(noteSchema).optional(),
    ClassTeacher: userSchema.nullable().optional(),
});

export type GettedUser = z.infer<typeof getUserResponseSchema>;

export const addUserDto = z.object({
    Name: z.string().min(2).max(20),
    LastName: z.string().min(2).max(20),
    FullName: z.array(fullNameSchema),
    Password: z.string().min(6).max(35),
    ClassID: z.number().int().nonnegative(),
    Login: z.string().min(2).max(20),
    Role: z.enum(["User", "Helper", "Admin", "Owner"]),
}).superRefine((data, ctx) => {
    const roleRequiresClass = data.Role === "User" || data.Role === "Helper";

    if (roleRequiresClass && data.ClassID <= 0) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: ["ClassID"],
            message: "Для ученика или старосты необходимо выбрать класс",
        });
    }
});

export type addUserType = z.infer<typeof addUserDto>;

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
    },
    
    async getStaff(): Promise<UserType[]> {
        const response = await client.get("/api/users/get/staff", true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении персонала");
        }
        
        const parsed = z.array(userSchema).safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }
        
        return parsed.data
    },
    
    async addUser(dto: addUserType): Promise<boolean> {
      const response = await client.patch("/api/user/add", {
          Name: dto.Name,
          LastName: dto.LastName,
          FullName: dto.FullName,
          Password: dto.Password,
          ClassID: dto.ClassID,
          Login: dto.Login,
          Role: dto.Role,
          Rating: 100,
      }, true);  
      
      if (!response.checkStatus()) {
          throw new Error("Ошибка при добавлении пользователя");
      }
      
      return true;
    }
}