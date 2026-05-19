import {z} from "zod";

export const ScheduleAddFormSchema = z.object({
    Subject: z.string().min(2).max(20),
    Room: z.number().int().positive(),
    StartTime: z.string().min(3).max(5),
    EndTime: z.string().min(3).max(5),
    TeacherID: z.number().int().positive(),
    LessonNumber: z.number().positive(),
    DayOfWeek: z.enum(["monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"]),
    ClassID: z.number().int().positive(),
});