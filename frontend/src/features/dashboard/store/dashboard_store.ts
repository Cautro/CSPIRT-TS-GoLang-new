import { create } from "zustand";
import {type addClassType, classApi} from "../../../shared/entities/class/api/class_api.ts";
import type {ClassType} from "../../../shared/entities/class/types/class_types.ts";
import {type addUserType, UserApi} from "../../../shared/entities/user/api/user_api.ts";
import type {UserType} from "../../../shared/entities/user/types/user_types.ts";
import type {EventType} from "../../../shared/entities/events/types/events_types.ts";
import {type addEventType, EventsApi} from "../../../shared/entities/events/api/events_api.ts";

export type DashboardStatus = "loading" | "error" | "idle";

interface State {
    status: DashboardStatus;
    error: string | null;
    message: string | null;
    classes: ClassType[];
    staff: UserType[];
    events: EventType[];

    getClasses: () => Promise<void>;
    addUser: (dto: addUserType) => Promise<void>;
    getStaff: () => Promise<void>;
    addClass: (dto: addClassType) => Promise<void>;
    getEvents: () => Promise<void>;
    addEvent: (dto: addEventType) => Promise<void>;
}

export const useDashboardStore = create<State>()((set) => ({
    error: null,
    status: "idle",
    message: null,
    classes: [],
    staff: [],
    events: [],
    
    getClasses: async () => {
      set({status: "loading"});
      
      try {
          const response = await classApi.getClasses();
          
          set({status: "idle", classes: response, error: null});
      } catch (e) {
          set({
              error: e instanceof Error ? e.message : "Неизвестная ошибка",
              status: "error",
          });
      }
    },
    
    addUser: async (dto: addUserType) => {
        set({status: "loading"});
        
        try {
            await UserApi.addUser(dto);
            set({status: "idle", message: "Пользователь успешно добавлен", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    getStaff: async () => {
        set({status: "loading"});
        
        try {
            const response = await UserApi.getStaff();
            set({staff: response, error: null, status: "idle"});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
    addClass: async (dto: addClassType) => {
        set({status: "loading"});
        
        try {
            await classApi.addClass(dto);
            set({status: "idle", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    },
    
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
    
    addEvent: async (dto: addEventType) => {
        set({status: "loading"});
        
        try {
            await EventsApi.addEvent(dto);
            set({status: "idle", message: "Мероприятие успешно добавлено", error: null});
        } catch (e) {
            set({
                error: e instanceof Error ? e.message : "Неизвестная ошибка",
                status: "error",
            });
        }
    }
    
}));