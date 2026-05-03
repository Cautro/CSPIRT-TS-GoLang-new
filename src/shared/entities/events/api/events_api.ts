import {ApiClient} from "../../../../core/api/api_client.ts";
import {EventSchema, type EventType} from "../types/events_types.ts";
import {z} from "zod";

const client = new ApiClient();

export const addEventSchema = z.object({
    Title: z.string(),
    Description: z.string(),
    StartedAt: z.string(),
    Classes: z.array(z.number().int().positive()),
    RatingReward: z.number().int().positive(),
});

export const addEventPlayersSchema = z.object({
   playerIds: z.array(z.number().int().positive()).min(1), 
});

export type AddEventPlayersType = z.infer<typeof addEventPlayersSchema>;


export type addEventType = z.infer<typeof addEventSchema>;

export const EventsApi = {
    async getEvents(): Promise<EventType[]> {
        const response = await client.get('/api/events', true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении списка мероприятий");
        }
        
        const parsed = z.array(EventSchema).safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        } 
        
        return parsed.data as EventType[];
    },
    
    async addEvent(dto: addEventType): Promise<boolean> {
        const response = await client.patch('/api/event/add', dto,true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при создании мероприятия");
        }
        
        return true;
    },

    async addPlayersToEvent(id: number, dto: AddEventPlayersType): Promise<boolean> {
        const response = await client.patch(`/api/event/${id}/players/add`, dto, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при добавлении участников мероприятия");
        }

        return true;
    },
    
    async removePlayersFromEvent(id: number, dto: AddEventPlayersType): Promise<boolean> {
        const response = await client.delete(`/api/event/${id}/players/delete`, dto, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при добавлении участников мероприятия");
        }

        return true;
    },
    
    async completeEvent(item: EventType): Promise<boolean> {
        const response = await client.patch(`/api/event/${item.ID}/complete`, {
            ratingReward: item.RatingReward
        }, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке завершения мероприятия");
        }
        
        return true;
    },
    
    async deleteEvent(id: number): Promise<boolean> {
        const response = await client.delete(`/api/event/delete/${id}`,{}, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при попытке удаления мероприятия");
        }

        return true;
    }
}