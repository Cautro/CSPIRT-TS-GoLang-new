import {useQuery} from "@tanstack/react-query";
import {UserApi} from "../../../shared/entities/user/api/user_api.ts";

export const useUser = (id: number) => {
    return useQuery({
       queryKey: ["user_by_id", id],
       queryFn: () => UserApi.getUser(id)
    }); 
}