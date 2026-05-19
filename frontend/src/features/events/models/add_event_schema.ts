import {z} from "zod";

export const addEventFormSchema = z.object({
    Title: z.string().min(4).max(64),
    Description: z.string().min(10).max(1000),
    StartedAt: z.string(),
    Classes: z.array(z.number().int().positive()),
    RatingReward: z.number().int().positive().max(5000),
});