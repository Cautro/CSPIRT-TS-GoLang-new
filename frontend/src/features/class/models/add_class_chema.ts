import {z} from "zod";

export const addClassFormSchema = z.object({
    Name: z.string().trim().min(1),
    TeacherLogin: z.string().trim().min(2),
});