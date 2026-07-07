import {useQuery} from "@tanstack/react-query";
import {EventsApi} from "../../../shared/entities/events/api/events_api.ts";

export const UseRewardParams = (id: number) => {
    return useQuery({
        queryKey: ["reward_params", id],
        queryFn: () => EventsApi.getRewardParams(id),
    })
}