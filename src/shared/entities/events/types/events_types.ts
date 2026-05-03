import {z} from "zod"

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