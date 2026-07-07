import {z} from "zod";

export const noteAddFormSchema = z.object({
    AuthorID: z.number().int().nonnegative(),
    CreatedAt: z.string(),
    TargetID: z.number().int().nonnegative(),
    Content: z.string().max(500),
    AuthorName: z.string(),
    TargetName: z.string(), 
});