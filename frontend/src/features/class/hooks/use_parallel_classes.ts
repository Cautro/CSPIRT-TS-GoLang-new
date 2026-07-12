import {useQuery} from "@tanstack/react-query";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";

export const useParallelClasses = (id: number) => {
    return useQuery({
        queryFn: () => classApi.getClassesByParallel(id),
        queryKey: ["parallel_classes", id]
    })
}