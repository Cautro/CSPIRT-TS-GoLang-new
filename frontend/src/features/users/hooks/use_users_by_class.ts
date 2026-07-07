import {useQuery} from "@tanstack/react-query";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";

export const UseUsersByClass = (id: number) => {
    return useQuery({
        queryKey: ["users_by_class", id],
        queryFn: () => classApi.getUsersByClass(id),
    })
}