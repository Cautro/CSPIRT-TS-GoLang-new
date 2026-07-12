import {useQuery} from "@tanstack/react-query";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";

export const useParallel = (id: number) => {
    return useQuery({
        queryFn: () => classApi.getParallelById(id),
        queryKey: ["parallel", id]
    })
}