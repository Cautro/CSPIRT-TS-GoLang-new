import {z} from "zod";

export const ScheduleChangeFormSchema = z.object({
    Subject: z.string().min(2).max(20).optional(),
    Room: z.number().int().positive().optional(),
    StartTime: z.string().min(3).max(5).optional(),
    EndTime: z.string().min(3).max(5).optional(),
    TeacherID: z.number().int().positive().optional(),
});