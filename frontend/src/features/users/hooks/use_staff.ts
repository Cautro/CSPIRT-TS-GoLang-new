import {useQuery} from "@tanstack/react-query";
import {UserApi} from "../../../shared/entities/user/api/user_api.ts";

export const useStaff = () => {
    return useQuery({
        queryKey: ["staff"],
        queryFn: () => UserApi.getStaff()
    })
}