import {useQuery} from "@tanstack/react-query";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";


export const useClasses = () => {
    return useQuery({
        queryKey: ["classes"],
        queryFn: () => classApi.getClasses(),
    });
}