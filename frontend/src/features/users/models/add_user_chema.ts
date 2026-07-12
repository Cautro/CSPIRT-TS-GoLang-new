import {fullNameSchema} from "../../../shared/entities/user/types/user_types.ts";
import {z} from "zod";

export const addUserFormSchema = z.object({
    Avatar: z.string(),
    Name: z.string().min(2).max(20),
    LastName: z.string().min(2).max(20),
    FullName: z.array(fullNameSchema).min(1),
    Password: z.string().min(6).max(35),
    ClassID: z.number().int().nonnegative(),
    Login: z.string().min(2).max(20),
    Role: z.enum(["User", "Helper", "Admin", "Owner", "Public"]),
    Rating: z.number().int().min(0).max(5000)
}).superRefine((data, ctx) => {
    const normalizedRole = data.Role.toLowerCase();
    const roleRequiresClass = normalizedRole === "user" || normalizedRole === "helper";

    if (roleRequiresClass && data.ClassID <= 0) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: ["ClassID"],
            message: "Для ученика или старосты необходимо выбрать класс",
        });
    }
});