import {ScheduleChangeFormSchema} from "./schedule_change_schema.ts";
import type {ScheduleChangeLessonFormType} from "../../../shared/entities/schedule/types/schedule_types.ts";

export type ScheduleChangeFormValues = {
    subject?: string,
    teacher_id?: number
    start_time?: string,
    end_time?: string,
    room?: number,
}

export function ScheduleChangeLessonUsecase(form: ScheduleChangeFormValues,): ScheduleChangeLessonFormType {
    const dto = {
        Subject: form.subject,
        TeacherID: form.teacher_id,
        StartTime: form.start_time,
        EndTime: form.end_time,
        Room: form.room,
    };

    const parsed = ScheduleChangeFormSchema.safeParse(dto);

    if (!parsed.success) {  
        throw new Error("Проверьте правильность заполнения полей");
    }
    

    return parsed.data as ScheduleChangeLessonFormType;
}