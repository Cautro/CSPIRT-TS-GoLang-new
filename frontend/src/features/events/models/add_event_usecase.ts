import {addEventFormSchema} from "./add_event_schema.ts";
import type {AddEventFormType} from "../../../shared/entities/events/types/events_types.ts";

export type addEventFormValues = {
    title: string;
    description: string;
    classes: number[];
    rating_reward: number;
    started_at: string;
}

export function addEventUsecase(form: addEventFormValues): AddEventFormType {
    const dto = {
        Title: form.title,
        Description: form.description,
        Classes: form.classes,
        RatingReward: form.rating_reward,
        StartedAt: form.started_at,
    };
    
    const parsed = addEventFormSchema.safeParse(dto);

    if (!parsed.success) {
        throw new Error("Проверьте правильность заполнения формы");
    }
    
    return parsed.data;
}