import {useMutation, useQueryClient} from "@tanstack/react-query";
import {type AddEventPlayersType, EventsApi} from "../../../shared/entities/events/api/events_api.ts";

export const useAddEventPlayers = () => {
    const queryclient= useQueryClient();
    
    return useMutation({
        mutationFn: ({dto, id}: {dto: AddEventPlayersType, id: number}) => EventsApi.addPlayersToEvent(id, dto),
        onSuccess: async () => {
            await queryclient.invalidateQueries({queryKey: ["events"]})
        }
    });
}