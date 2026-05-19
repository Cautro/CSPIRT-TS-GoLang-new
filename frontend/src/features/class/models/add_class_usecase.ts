import type {addClassFormType} from "../../../shared/entities/class/types/class_types.ts";
import {addClassFormSchema} from "./add_class_chema.ts";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";

export type addClassFormValues = {
    name: string;
    teacher_login: string;
}

export async function AddClassUsecase(form: addClassFormValues): Promise<boolean> {
    const dto: addClassFormType = {
        Name: form.name,
        TeacherLogin: form.teacher_login,
    };
    
    const parsed = addClassFormSchema.safeParse(dto);
    
    if (!parsed.success) {
        throw new Error(JSON.stringify(parsed.error?.format()));
    }
    
    await classApi.addClass(parsed.data);
    
    return true
}