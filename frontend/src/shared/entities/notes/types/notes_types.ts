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

export type NoteFormType = {
    AuthorID: number;
    TargetID: number;
    Content: string;
    CreatedAt: string;
    AuthorName: string;
    TargetName: string;
}