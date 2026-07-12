import type {UserRole} from "../../../shared/entities/user/types/user_types.ts";
import {addUserFormSchema} from "./add_user_chema.ts";
import {UserApi} from "../../../shared/entities/user/api/user_api.ts";

export type addUserValues = {
    avatar: string;
    name: string;
    lastname: string;
    password: string;
    classId: number;
    login: string;
    role: UserRole;
}

export async function AddUserUseCase(form: addUserValues): Promise<boolean> {
    
    const dto = {
        Avatar: form.avatar,
        Name: form.name,
        LastName: form.lastname,
        Password: form.password,
        ClassID: form.classId,
        Login: form.login,
        Role: form.role,
        FullName: [{
            Name: form.name,
            LastName: form.lastname,
        }],
        Rating: 100
    }
    const parsed = addUserFormSchema.safeParse(dto);
    
    if (!parsed.success) {
        throw new Error("Проверьте правильность заполнения полей");
    }
    
    await UserApi.addUser(parsed.data);
    
    return true;
}