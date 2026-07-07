import {useMutation, useQueryClient} from "@tanstack/react-query";
import {ScheduleApi} from "../../../shared/entities/schedule/api/schedule_api.ts";

export const UseRolloverSchedule = () => {
    const queryclient = useQueryClient();

    return useMutation({
        mutationFn: ({id}: {id: number}) => ScheduleApi.rolloverSchedule(id),
        onSuccess: () => {
            queryclient.invalidateQueries({queryKey: ["class_schedule"]});
            queryclient.invalidateQueries({queryKey: ["teacher_schedule"]});
        }
    });
}