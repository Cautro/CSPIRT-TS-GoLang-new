import {z} from "zod";

export const EventSchema = z.object({
    ID: z.number().int().positive(),
    Title: z.string(),
    Status: z.string(),
    RatingReward: z.number().int().nonnegative(),
    Description: z.string(),
    CreatedAt: z.string(),
    StartedAt: z.string(),
    Players: z.array(z.number().int().positive()),
    Classes: z.array(z.number().int().positive()),
});

export type EventType = z.infer<typeof EventSchema>;

export type AddEventFormType = {
    Title: string;
    Description: string;
    StartedAt: string;
    Classes: number[];
    RatingReward: number;
}
    
export type AddRewardParamsFormType = {
    ClassID: number;
    ExtraRatingReward: number;
    Reason: string;
}

export const RewardParamsSchema = z.object({
    ExtraRatingReward: z.number().int().positive(),
    Reason: z.string(),
    ClassID: z.number().int().nonnegative(),
})

export type RewardParamsType = z.infer<typeof RewardParamsSchema>;