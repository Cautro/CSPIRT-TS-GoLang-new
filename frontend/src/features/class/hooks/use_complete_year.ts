import {useMutation, useQueryClient} from "@tanstack/react-query";
import {classApi} from "../../../shared/entities/class/api/class_api.ts";

export const useCompleteYear = () => {
    const queryclient = useQueryClient();

    return useMutation({
        mutationFn: () =>
            classApi.completeYear().then((result) => {return result}),
        onSuccess: () => {
            queryclient.invalidateQueries({queryKey: ["classes"]});
            queryclient.invalidateQueries({queryKey: ["parallels"]});
            queryclient.invalidateQueries({queryKey: ["parallel_classes"]});
            queryclient.invalidateQueries({queryKey: ["parallel"]});
            queryclient.invalidateQueries({queryKey: ["class_by_id"]});
        },
    })
}