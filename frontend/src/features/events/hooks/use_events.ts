import {useQuery} from "@tanstack/react-query";
import {EventsApi} from "../../../shared/entities/events/api/events_api.ts";

export const UseEvents = () => {
    return useQuery({
        queryKey: ["events"],
        queryFn: () => EventsApi.getEvents()
    })
}