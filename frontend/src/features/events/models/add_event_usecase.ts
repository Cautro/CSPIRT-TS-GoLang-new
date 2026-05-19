import {addEventFormSchema} from "./add_event_schema.ts";
import {EventsApi} from "../../../shared/entities/events/api/events_api.ts";

export type addEventFormValues = {
    title: string;
    description: string;
    classes: number[];
    rating_reward: number;
    started_at: string;
}

export async function addEventUsecase(form: addEventFormValues): Promise<boolean> {
    const dto = {
        Title: form.title,
        Description: form.description,
        Classes: form.classes,
        RatingReward: form.rating_reward,
        StartedAt: form.started_at,
    };
    
    const parsed = addEventFormSchema.safeParse(dto);
    
    if (!parsed.success) {
        throw new Error(JSON.stringify(parsed.error?.format()));
    } 
    
    await EventsApi.addEvent(parsed.data);
    
    return true;
}