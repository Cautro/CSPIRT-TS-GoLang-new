import {useMutation, useQueryClient} from "@tanstack/react-query";
import {ScheduleApi} from "../../../shared/entities/schedule/api/schedule_api.ts";
import type {ScheduleChangeLessonFormType} from "../../../shared/entities/schedule/types/schedule_types.ts";

export const UseChangeSchedule = () => {
    const queryclient = useQueryClient();
    
    return useMutation({
        mutationFn: ({id, form, type}: {id: number, form: ScheduleChangeLessonFormType, type: "base" | "current" | "planned"}) => ScheduleApi.changeScheduleLesson(id, form, type),
        onSuccess: () => {
            queryclient.invalidateQueries({queryKey: ["class_schedule"]});
            queryclient.invalidateQueries({queryKey: ["teacher_schedule"]});
        }
    });
}