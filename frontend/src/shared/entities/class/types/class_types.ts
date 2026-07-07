import {z} from "zod";
import {userSchema} from "../../user/types/user_types.ts";

export const classSchema = z.object({
    Id: z.number().nonnegative(),
    Name: z.string(),
    TeacherLogin: z.string().optional(),
    Members: z.array(userSchema),
    Teacher: userSchema.optional(),
    UserTotalRating: z.number(),
    ClassTotalRating: z.number(),
});

export type ClassType = z.infer<typeof classSchema>;

export type addClassFormType = {
    Name: string;
    TeacherLogin: string;
}