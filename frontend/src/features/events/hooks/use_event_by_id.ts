import {useQuery} from "@tanstack/react-query";
import {EventsApi} from "../../../shared/entities/events/api/events_api.ts";

export const UseEventById = (id: number) => {
    return useQuery({
        queryKey: ["event_by_id", id],
        queryFn: () => EventsApi.getEventById(id)
    })
}