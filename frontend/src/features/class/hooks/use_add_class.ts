import {useMutation, useQueryClient} from "@tanstack/react-query";
import {type addClassFormValues, AddClassUsecase} from "../models/add_class_usecase.ts";


export const useAddClass = () => {
    const queryClient = useQueryClient();
    
    return useMutation({
        mutationFn: (form: addClassFormValues) => AddClassUsecase(form),
        onSuccess: async () => {
            await queryClient.invalidateQueries({queryKey: ["classes"]});
        }
    });
}