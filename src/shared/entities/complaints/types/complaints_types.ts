import {z} from "zod";

export const complaintSchema = z.object({
    ID: z.number().nonnegative(),
    AuthorID: z.number(),
    TargetID: z.number(),
    Content: z.string(),
    CreatedAt: z.string(),
    AuthorName: z.string(),
    TargetName: z.string(),
});

export type ComplaintType = z.infer<typeof complaintSchema>;