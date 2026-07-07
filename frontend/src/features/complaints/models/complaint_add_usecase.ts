import {complaintAddFormSchema} from "./complaint_add_schema.ts";
import type {UserType} from "../../../shared/entities/user/types/user_types.ts";
import type {complaintAddFormType} from "../../../shared/entities/complaints/api/complaints_api.ts";

export interface ComplaintAddFormValues {
    user: UserType,
    current_user: UserType,
    content: string,
}

export const ComplaintAddUsecase = (form: ComplaintAddFormValues): complaintAddFormType => {

    if (!form.user || !form.current_user) {
        throw new Error("Не удалось определить пользователя");
    }

    const dto = {
        TargetID: form.user.Id,
        Content: form.content,
        AuthorID: form.current_user.Id,
        CreatedAt: new Date().toISOString(),
        AuthorName: `${form.current_user.Name} ${form.current_user.LastName}`,
        TargetName: `${form.user.Name} ${form.user.LastName}`,
    };

    const parsed = complaintAddFormSchema.safeParse(dto);

    if (!parsed.success) {
        throw new Error("Проверьте текст жалобы");
    }
    
    return parsed.data as complaintAddFormType;
}