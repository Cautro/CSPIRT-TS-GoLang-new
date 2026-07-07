import {RatingChangeFormSchema} from "./rating_change_schema.ts";
import type {RatingChangeFormType} from "../../../shared/entities/rating/types/rating_types.ts";

export interface RatingChangeFormValues {
    target: string;
    reason: string;
    rating: number;
}

export const RatingChangeUseCase = (form:  RatingChangeFormValues): RatingChangeFormType => {
    const dto = {
        target_login: form.target,
        reason: form.reason,
        rating: form.rating,
    };
    
    const parsed = RatingChangeFormSchema.safeParse(dto);
    
    if (!parsed.success) {
        throw new Error('Проверьте правильность заполнения формы');
    }
    
    return parsed.data;
}