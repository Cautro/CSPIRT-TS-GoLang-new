import {z} from "zod";

export const noteSchema = z.object({
    ID: z.number(),
    AuthorID: z.number(),
    TargetID: z.number(),
    Content: z.string(),
    CreatedAt: z.string(),
});

export type NoteType = z.infer<typeof noteSchema>;