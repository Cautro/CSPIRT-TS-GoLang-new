import {z} from "zod";

export const AddRewardParamsFormSchema = z.object({
   ClassID: z.number().nonnegative(), 
   ExtraRatingReward: z.number().positive().max(5000),
   Reason: z.string().max(50),
});