import {useMutation, useQueryClient} from "@tanstack/react-query";
import {UserApi} from "../../../shared/entities/user/api/user_api.ts";

export const useDeleteUser = () => {
    const queryclient = useQueryClient();
    
    return useMutation({
        mutationFn: ({id}: {id: number}) => UserApi.deleteUser(id),
        onSuccess: () => {
            queryclient.invalidateQueries({queryKey: ["users_by_class",]});
        }
    })
}