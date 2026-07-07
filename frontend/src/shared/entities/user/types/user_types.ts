import { z } from "zod";

const nullableArray = <T extends z.ZodTypeAny>(schema: T) =>
    z.array(schema)
        .nullish()
        .transform((value) => value ?? []);

const roles = ["Owner", "Admin", "Helper", "User", "Public"] as const;

export const userRoleSchema = z
    .string()
    .transform((value) => {
        const lower = value.toLowerCase();

        switch (lower) {
            case "owner":
                return "Owner";
            case "admin":
                return "Admin";
            case "helper":
                return "Helper";
            case "user":
                return "User";
            case "public":
                return "Public";
            default:
                return value;
        }
    })
    .pipe(z.enum(roles));

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

export type UserType = z.infer<typeof userSchema>;
export type UserRole = z.infer<typeof userRoleSchema>;

export const UserRoles: Record<UserRole, string> = {
    Admin: "Учитель",
    User: "Ученик",
    Owner: "Руководство",
    Helper: "Староста",
    Public: "Публичный"
};

export type addUserFormType = {
    Name: string;
    LastName: string;
    FullName: Array<{
        Name: string;
        LastName: string;
    }>
    Password: string;
    ClassID: number;
    Login: string;
    Role: UserRole;
    Rating: number;
}