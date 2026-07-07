import {useMutation, useQueryClient} from "@tanstack/react-query";
import type {EventType} from "../../../shared/entities/events/types/events_types.ts";
import {EventsApi} from "../../../shared/entities/events/api/events_api.ts";

export const useCompleteEvent = () => {
    const queryclient= useQueryClient();
    
    return useMutation({
        mutationFn: ({item}: {item: EventType}) => EventsApi.completeEvent(item),
        onSuccess: async () => {
            await queryclient.invalidateQueries({queryKey: ["events"]});
            await queryclient.invalidateQueries({queryKey: ["classes"]});
        }
    })
}