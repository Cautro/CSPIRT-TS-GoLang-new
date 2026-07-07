import {useMutation, useQueryClient} from "@tanstack/react-query";
import {ScheduleApi} from "../../../shared/entities/schedule/api/schedule_api.ts";

export const UseDeleteSchedule = () => {
    const queryclient = useQueryClient();

    return useMutation({
        mutationFn: ({id, type}: {id: number, type: "base" | "current" | "planned"}) => ScheduleApi.deleteSchedule(id, type),
        onSuccess: () => {
            queryclient.invalidateQueries({queryKey: ["class_schedule"]});
            queryclient.invalidateQueries({queryKey: ["teacher_schedule"]});
        }
    });
}