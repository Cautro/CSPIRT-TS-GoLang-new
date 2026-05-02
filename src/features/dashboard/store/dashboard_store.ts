import { create } from "zustand";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";
import type {ClassType} from "../../../shared/entities/class/types/class_types.ts";

export type DashboardStatus = "loading" | "error" | "idle";

interface State {
    status: DashboardStatus;
    error: string | null;
    message: string | null;
    classes: ClassType[];

    getClasses: () => Promise<void>
}

export const useDashboardStore = create<State>()((set) => ({
    error: null,
    status: "idle",
    message: null,
    classes: [],
    
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
    
    
}));