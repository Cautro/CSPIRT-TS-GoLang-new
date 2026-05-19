import {userSchema} from "../../user/types/user_types.ts";
import {z} from "zod";


export const ScheduleLessonSchema = z.object({
    Id: z.number().int(),
    Type: z.enum(["base", "current", "planned"]).optional(),
    BaseScheduleID: z.number().int().optional(),
    ClassID: z.number().int(),
    Class: z.string().optional(),
    DayOfWeek: z.string(),
    LessonNumber: z.number().int().min(1),
    WeekType: z.union([
        z.enum(["all", "odd", "even"]),
        z.string()
    ]),
    Subject: z.string().min(1, "Название предмета обязательно"),
    TeacherID: z.number().int(),
    Teacher: userSchema.optional(),
    Room: z.number().int(),
    StartTime: z.string().regex(/^([01]\d|2[0-3]):([0-5]\d)$/, "Формат времени HH:mm"),
    EndTime: z.string().regex(/^([01]\d|2[0-3]):([0-5]\d)$/, "Формат времени HH:mm"),
    Description: z.string(),
    CreatedAt: z.string().optional(),
});

export type ScheduleLessonType = z.infer<typeof ScheduleLessonSchema>;

export const SchedulesResponseSchema = z.object({
    Schedules: z.array(ScheduleLessonSchema),
    Base: z.array(ScheduleLessonSchema),
    Current: z.array(ScheduleLessonSchema),
    Planned: z.array(ScheduleLessonSchema),
});

export const TeacherScheduleResponseSchema = z.object({
    Schedules: z.array(ScheduleLessonSchema),
});

export const dayLabels = {
    monday: "Понедельник",
    tuesday: "Вторник",
    wednesday: "Среда",
    thursday: "Четверг",
    friday: "Пятница",
    saturday: "Суббота",
    sunday: "Воскресенье",
};

export type ScheduleModelType = {
    monday: ScheduleLessonType[],
    tuesday: ScheduleLessonType[],
    wednesday: ScheduleLessonType[],
    thursday: ScheduleLessonType[],
    friday: ScheduleLessonType[],
    saturday: ScheduleLessonType[],
    sunday: ScheduleLessonType[],
}

export type ScheduleChangeLessonFormType = {
    Subject?: string,
    Room?: number,
    StartTime?: string,
    EndTime?: string,
    TeacherID?: number,
}

export type ScheduleAddLessonFormType = {
    ClassID: number,
    DayOfWeek: string,
    Subject: string,
    TeacherID: number,
    Room: number,
    StartTime: string,
    EndTime: string,
    LessonNumber: number,
}

export type ScheduleDay =
    | "monday"
    | "tuesday"
    | "wednesday"
    | "thursday"
    | "friday"
    | "saturday"
    | "sunday";