import {useMutation, useQueryClient} from "@tanstack/react-query";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";

export const useDeleteClass = () => {
    const queryclient = useQueryClient();
    
    return useMutation({
        mutationFn: ({id}: {id: number}) => classApi.deleteClass(id),
        onSuccess: (_data, variables) => {
            const { id } = variables;
            
            queryclient.invalidateQueries({queryKey: ["classes", id]});
            queryclient.invalidateQueries({queryKey: ["class_by_id", id]});
        }
    })
}