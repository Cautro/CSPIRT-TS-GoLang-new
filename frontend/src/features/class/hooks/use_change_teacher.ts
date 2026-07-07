import {useMutation, useQueryClient} from "@tanstack/react-query";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";

export const useChangeTeacher = () => {
    const queryClient = useQueryClient();
    
    return useMutation({
        mutationFn: ({id, teacher}: {id: number, teacher: string}) => classApi.changeClassTeacher(id, teacher),
        onSuccess: (_data, variables) => {
            const { id } = variables;
            
            queryClient.invalidateQueries({queryKey: ["classes", id]});
            queryClient.invalidateQueries({queryKey: ["class_by_id", id]});
        }
    })
}