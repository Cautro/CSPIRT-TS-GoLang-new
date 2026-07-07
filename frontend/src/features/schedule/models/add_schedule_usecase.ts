import {ScheduleAddFormSchema} from "./add_schedule_schema.ts";
import type {ScheduleAddLessonFormType} from "../../../shared/entities/schedule/types/schedule_types.ts";

export type ScheduleAddFormValues = {
    subject: string,
    teacher_id: number
    start_time: string,
    end_time: string,
    room: number,
    day_of_week: string,
    class_id: number,
    lesson_number: number,
}

export function ScheduleAddLessonUsecase(form: ScheduleAddFormValues): ScheduleAddLessonFormType {
    const dto = {
        Subject: form.subject,
        TeacherID: form.teacher_id,
        StartTime: form.start_time,
        EndTime: form.end_time,
        Room: form.room,
        ClassID: form.class_id,
        DayOfWeek: form.day_of_week,
        LessonNumber: form.lesson_number,
    };

    const parsed = ScheduleAddFormSchema.safeParse(dto);

    if (!parsed.success) {
        throw new Error("Проверьте правильность заполнения полей");
    }

    return parsed.data;
}