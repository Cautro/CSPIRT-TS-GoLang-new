import type { UserType } from "../../../shared/entities/user/types/user_types.ts";
import type { ClassType } from "../../../shared/entities/class/types/class_types.ts";
import { create } from "zustand";
import { classApi } from "../../../shared/entities/class/api/class_api.ts";
import {
    type AddEventPlayersType,
    EventsApi,
} from "../../../shared/entities/events/api/events_api.ts";
import type { EventType } from "../../../shared/entities/events/types/events_types.ts";

export type Status = "loading" | "error" | "idle";

interface State {
    status: Status;
    error: string | null;
    message: string | null;

    users: UserType[];
    classes: ClassType[];
    event: EventType | null;

    getClasses: () => Promise<void>;
    removePlayersFromEvent : (id: number, dto: AddEventPlayersType) => Promise<void>; 
    addPlayersToEvent: (
        eventId: number,
        dto: AddEventPlayersType
    ) => Promise<void>;
    getEventById: (id: number) => Promise<void>;
    completeEvent: (item: EventType) => Promise<void>;
    deleteEvent: (id: number) => Promise<void>;
}

export const useEventStore = create<State>()((set) => ({
    error: null,
    status: "idle",
    message: null,

    users: [],
    classes: [],
    event: null,

    getClasses: async () => {
        set({
            status: "loading",
            error: null,
        });

        try {
            const response = await classApi.getClasses();

            set({
                status: "idle",
                classes: response,
                error: null,
            });
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
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
            const response = await EventsApi.getEvents();

            const foundEvent = response.find((item) => item.ID === id) ?? null;

            set({status: "idle", event: foundEvent, error: null,});
            
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
                event: null,
            });
        }
    },
    
    completeEvent: async (item: EventType) => {
        set({status: "loading", error: null, event: null,});

        try {
            await EventsApi.completeEvent(item);
            set({status: "idle", event: null, message: "Мероприятие успешно завершено"});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
                event: null,
            });
        }
    },
    
    deleteEvent: async (id: number) => {
        set({status: "loading", error: null, event: null,});

        try {
            await EventsApi.deleteEvent(id);
            set({status: "idle", event: null, message: "Мероприятие успешно удалено"});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
                event: null,
            });
        }
    }
}));