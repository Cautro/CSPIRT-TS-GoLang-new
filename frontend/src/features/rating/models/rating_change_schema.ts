import {z} from "zod";

export const RatingChangeFormSchema = z.object({
    rating: z.number().int().max(5000).min(-5000),
    target_login: z.string(),
    reason: z.string(),
});