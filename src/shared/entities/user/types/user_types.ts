import { z } from "zod";

const nullableArray = <T extends z.ZodTypeAny>(schema: T) =>
    z.array(schema)
        .nullish()
        .transform((value) => value ?? []);

export const userRoleSchema = z.enum(["Owner", "Admin", "Helper", "User"]);

export const fullNameSchema = z.object({
    Name: z.string(),
    LastName: z.string(),
});

export const userSchema = z.object({
    Id: z.number().int().nonnegative(),
    Name: z.string().max(100),
    LastName: z.string().max(100),

    FullName: nullableArray(fullNameSchema),

    Login: z.string().min(1).max(64),
    Rating: z.number().int(),
    Role: userRoleSchema,
    Class: z.string().max(32),
    ClassID: z.number().int().nonnegative(),
});

export const usersSchema = z.array(userSchema);

export type UserType = z.infer<typeof userSchema>;
export type UserRole = z.infer<typeof userRoleSchema>;

export const UserRoles: Record<UserRole, string> = {
    Admin: "Учитель",
    User: "Ученик",
    Owner: "...",
    Helper: "Староста",
};