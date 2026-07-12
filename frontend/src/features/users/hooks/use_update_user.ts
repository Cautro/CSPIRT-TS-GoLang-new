import {useMutation, useQueryClient} from "@tanstack/react-query";
import {UpdateUserUseCase, type updateUserValues} from "../models/update_user_usecase.ts";

export const useUpdateUser = () => {
    const queryclient = useQueryClient();

    return useMutation({
        mutationFn: (form: updateUserValues) => UpdateUserUseCase(form),
        onSuccess: async () => {
            await queryclient.invalidateQueries({queryKey: ["user_by_id"]});
            await queryclient.invalidateQueries({queryKey: ["users_by_class"]});
            await queryclient.invalidateQueries({queryKey: ["staff"]});
        }
    })
}