import {useMutation, useQueryClient} from "@tanstack/react-query";
import {ScheduleApi} from "../../../shared/entities/schedule/api/schedule_api.ts";
import type {ScheduleAddLessonFormType,} from "../../../shared/entities/schedule/types/schedule_types.ts";

export const UseAddSchedule = () => {
    const queryclient = useQueryClient();

    return useMutation({
        mutationFn: ({form, type}: {form: ScheduleAddLessonFormType, type: "base" | "current" | "planned"}) => ScheduleApi.addScheduleLesson(form, type),
        onSuccess: () => {
            queryclient.invalidateQueries({queryKey: ["class_schedule"]});
            queryclient.invalidateQueries({queryKey: ["teacher_schedule"]});
        }
    });
}