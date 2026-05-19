import {ScheduleChangeFormSchema} from "./schedule_change_schema.ts";
import {ScheduleApi} from "../../../shared/entities/schedule/api/schedule_api.ts";

export type ScheduleChangeFormValues = {
    subject?: string,
    teacher_id?: number
    start_time?: string,
    end_time?: string,
    room?: number,
}

export async function ScheduleChangeLessonUsecase(id: number, form: ScheduleChangeFormValues, type: "base" | "current" | "planned"): Promise<boolean> {
    const dto = {
        Subject: form.subject,
        TeacherID: form.teacher_id,
        StartTime: form.start_time,
        EndTime: form.end_time,
        Room: form.room,
    };

    const parsed = ScheduleChangeFormSchema.safeParse(dto);

    if (!parsed.success) {  
        throw new Error(JSON.stringify(parsed.error?.format()));
    }

    await ScheduleApi.changeScheduleLesson(id, parsed.data, type);

    return true;
}