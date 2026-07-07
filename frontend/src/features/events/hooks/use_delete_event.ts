import {useMutation, useQueryClient} from "@tanstack/react-query";
import {EventsApi} from "../../../shared/entities/events/api/events_api.ts";

export const useDeleteEvent = () => {
    const queryclient= useQueryClient();

    return useMutation({
        mutationFn: ({id}: {id: number}) => EventsApi.deleteEvent(id),
        onSuccess: async () => {
            await queryclient.invalidateQueries({queryKey: ["events"]});
        }
    })
}