import {useQuery} from "@tanstack/react-query";
import {ScheduleApi} from "../../../shared/entities/schedule/api/schedule_api.ts";
import {ScheduleService} from "../../../shared/entities/schedule/service/schedule_service.ts";

export const useClassSchedule = (id: number, type: "base" | "current" | "planned") => {
    return useQuery({
        queryKey: ["class_schedule", id],
        queryFn: async () => {
            const response = await ScheduleApi.getCurrentScheduleByClass(id, type);
            return ScheduleService.sortSchedule(response);
        }
    });
}