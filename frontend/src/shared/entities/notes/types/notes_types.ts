import {z} from "zod";

export const noteSchema = z.object({
    ID: z.number().nonnegative(),
    AuthorID: z.number(),
    TargetID: z.number(),
    Content: z.string(),
    CreatedAt: z.string(),
    AuthorName: z.string(),
    TargetName: z.string(),
});

export type NoteType = z.infer<typeof noteSchema>;