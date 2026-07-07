import {useMutation, useQueryClient} from "@tanstack/react-query";
import {AddUserUseCase, type addUserValues} from "../models/add_user_usecase.ts";

export const useAddUser = () => {
    const queryclient = useQueryClient();
    
    return useMutation({
        mutationFn: (form: addUserValues) => AddUserUseCase(form),
        onSuccess: async () => {
            await queryclient.invalidateQueries({queryKey: ["users_by_class"]});
            await queryclient.invalidateQueries({queryKey: ["staff"]});
        }
    })
}