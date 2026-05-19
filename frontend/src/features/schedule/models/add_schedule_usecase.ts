import {ScheduleApi} from "../../../shared/entities/schedule/api/schedule_api.ts";
import {ScheduleAddFormSchema} from "./add_schedule_schema.ts";

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

export async function ScheduleAddLessonUsecase(form: ScheduleAddFormValues, type: "base" | "current" | "planned"): Promise<boolean> {
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
        throw new Error(JSON.stringify(parsed.error?.format()));
    }

    await ScheduleApi.addScheduleLesson(parsed.data, type);

    return true;
}