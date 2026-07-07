import {useMutation} from "@tanstack/react-query";
import {EventsApi} from "../../../shared/entities/events/api/events_api.ts";
import type {AddEventFormType} from "../../../shared/entities/events/types/events_types.ts";

export const useAddEvent = () => {
    return useMutation({
        mutationFn: (form: AddEventFormType) => EventsApi.addEvent(form)
    })
}