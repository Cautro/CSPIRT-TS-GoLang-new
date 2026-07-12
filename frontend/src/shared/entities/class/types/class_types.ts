import {z} from "zod";
import {userSchema} from "../../user/types/user_types.ts";

export const classSchema = z.object({
    Id: z.number().nonnegative(),
    Name: z.string(),
    Grade: z.number().nonnegative(),
    Letter: z.string(),
    FirstQuarterComplete: z.number().nonnegative(),
    SecondQuarterComplete: z.number().nonnegative(),
    ThirdQuarterComplete: z.number().nonnegative(),
    QuarterComplete: z.number().nonnegative(),
    TeacherLogin: z.string().optional(),
    Members: z.array(userSchema),
    Teacher: userSchema.optional(),
    UserTotalRating: z.number(),
    ClassTotalRating: z.number(),
});

export const parallelSchema = z.object({
    Id: z.number().nonnegative(),
    Name: z.string(),
    BestClassId: z.number().nonnegative(),
    ClassesIds: z.array(z.number().nonnegative()).nullish(),
})

export const quarterCompleteSchema = z.object({
    "1st": classSchema.nullable(),
    "2nd": classSchema.nullable(),
    "3rd": classSchema.nullable(),
    message: z.string()
})

export type QuarterCompleteType = z.infer<typeof quarterCompleteSchema>


export type ParallelType = z.infer<typeof parallelSchema>

export type ClassType = z.infer<typeof classSchema>;

export type addClassFormType = {
    Name: string;
    TeacherLogin: string;
}