import {useMutation, useQueryClient} from "@tanstack/react-query";
import {type AddParamFormValues, AddParamUsecase} from "../models/add_param_usecase.ts";

export const UseAddParam = () => {
    const queryclient = useQueryClient();
    
    return useMutation({
        mutationFn: async (form: AddParamFormValues) => AddParamUsecase(form),
        onSuccess: () =>  {
            queryclient.invalidateQueries({queryKey: ["reward_params"]});
        }
    })
}