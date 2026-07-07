import {
    type AddEventFormType,
    type AddRewardParamsFormType,
    EventSchema,
    type EventType, RewardParamsSchema, type RewardParamsType
} from "../types/events_types.ts";
import {z} from "zod";
import {apiClient} from "../../../../core/api/client.ts";

export const addEventPlayersSchema = z.object({
   playerIds: z.array(z.number().int().positive()).min(1), 
});

export type AddEventPlayersType = z.infer<typeof addEventPlayersSchema>;

export const EventsApi = {
    async getEvents(): Promise<EventType[]> {
        const response = await apiClient.get('/api/events', true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении списка мероприятий");
        }
        
        const parsed = z.array(EventSchema).safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        } 
        
        return parsed.data;
    },
    
    async getEventById(id: number): Promise<EventType> {
        const response = await apiClient.get(`/api/events?event_id=${id}`, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении мероприятия");
        }

        const parsed =EventSchema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }
        
        return parsed.data;
    },
    
    async addEvent(dto: AddEventFormType): Promise<boolean> {
        const response = await apiClient.patch('/api/event/add', dto,true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при создании мероприятия");
        }
        
        return true;
    },

    async addPlayersToEvent(id: number, dto: AddEventPlayersType): Promise<boolean> {
        const response = await apiClient.patch(`/api/event/${id}/players/add`, dto, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при добавлении участников мероприятия");
        }

        return true;
    },
    
    async removePlayersFromEvent(id: number, dto: AddEventPlayersType): Promise<boolean> {
        const response = await apiClient.delete(`/api/event/${id}/players/delete`, dto, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при удалении участников мероприятия");
        }

        return true;
    },
    
    async completeEvent(item: EventType): Promise<boolean> {
        const response = await apiClient.patch(`/api/event/${item.ID}/complete`, {
            ratingReward: item.RatingReward
        }, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке завершения мероприятия");
        }
        
        return true;
    },
    
    async deleteEvent(id: number): Promise<boolean> {
        const response = await apiClient.delete(`/api/event/delete/${id}`,{}, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке удаления мероприятия");
        }

        return true;
    },
    
    async getRewardParams(id: number): Promise<RewardParamsType[]> {
        const response = await apiClient.get(`/api/event/${id}/params`);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке получения параметров награждения");
        }
        
        const parsed = z.array(RewardParamsSchema).safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }
        
        return parsed.data;
    },
    
    async addRewardParams(id: number, form: AddRewardParamsFormType) {
        const response = await apiClient.patch(`/api/event/${id}/params/add`, form, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке добавления параметров награждения");
        }
    }
}