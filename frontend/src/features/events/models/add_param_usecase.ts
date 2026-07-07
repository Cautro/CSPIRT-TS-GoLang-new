import {AddRewardParamsFormSchema} from "./add_param_schema.ts";
import {EventsApi} from "../../../shared/entities/events/api/events_api.ts";

export type AddParamFormValues = {
    class_id: number;
    reason: string;
    event_id: number;
    rating: number;
}

export async function AddParamUsecase(form: AddParamFormValues) {
    const dto = {
        ClassID: form.class_id,
        Reason: form.reason,
        ExtraRatingReward: form.rating
    };

    const parsed = AddRewardParamsFormSchema.safeParse(dto);

    if (!parsed.success) {
        throw new Error("Проверьте правильность заполнения формы");
    }
    
    await EventsApi.addRewardParams(form.event_id, dto);
}