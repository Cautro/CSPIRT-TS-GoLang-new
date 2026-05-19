import { create } from "zustand";
import {
    type AddEventPlayersType,
    EventsApi,
} from "../../../shared/entities/events/api/events_api.ts";
import type { EventType } from "../../../shared/entities/events/types/events_types.ts";
import {type addEventFormValues, addEventUsecase} from "../models/add_event_usecase.ts";

export type Status = "loading" | "error" | "idle";

interface State {
    status: Status;
    error: string | null;
    message: string | null;
    event: EventType | null;
    events: EventType[] | null;

    getEvents: () => Promise<void>
    addEvent: (form: addEventFormValues) => Promise<boolean>
    removePlayersFromEvent : (id: number, dto: AddEventPlayersType) => Promise<void>; 
    addPlayersToEvent: (eventId: number, dto: AddEventPlayersType) => Promise<void>;
    getEventById: (id: number) => Promise<void>;
    completeEvent: (item: EventType) => Promise<void>;
    deleteEvent: (id: number) => Promise<void>;
}

export const useEventStore = create<State>()((set) => ({
    error: null,
    status: "idle",
    message: null,
    event: null,
    events: null,

    getEvents: async () => {
        set({status: "loading"});

        try {
            const response = await EventsApi.getEvents();
            set({status: "idle", events: response, error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },

    addEvent: async (form: addEventFormValues) => {
        set({status: "loading", error: null});

        try {
            const response = await addEventUsecase(form);
            if (response) {
                set({status: "idle", error: null});
                return true;
            }
            return false;
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
            return false
        }
    },

    addPlayersToEvent: async (eventId, dto) => {
        set({status: "loading", error: null, message: null,});

        try {
            await EventsApi.addPlayersToEvent(eventId, dto);

            set({status: "idle", error: null, message: "Участники успешно добавлены",
            });
        } catch (e) {
            set({
                status: "error",
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
            });
        }
    },

    removePlayersFromEvent: async (eventId, dto) => {
        set({status: "loading", error: null, message: null,});

        try {
            await EventsApi.removePlayersFromEvent(eventId, dto);

            set({status: "idle", error: null, message: "Участники успешно удалены",
            });
        } catch (e) {
            set({
                status: "error",
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
            });
        }
    },

    getEventById: async (id: number) => {
        set({status: "loading", error: null, event: null,});

        try {
            const response = await EventsApi.getEventById(id);
            set({status: "idle", event: response, error: null,});
            
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
                event: null,
            });
        }
    },
    
    completeEvent: async (item: EventType) => {
        set({status: "loading", error: null,});

        try {
            await EventsApi.completeEvent(item);
            set({status: "idle", event: null, message: "Мероприятие успешно завершено", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
                event: null,
            });
        }
    },
    
    deleteEvent: async (id: number) => {
        set({status: "loading", error: null,});

        try {
            await EventsApi.deleteEvent(id);
            set({status: "idle", event: null, message: "Мероприятие успешно удалено", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
                event: null,
            });
        }
    },
}));