import {useMutation, useQueryClient} from "@tanstack/react-query";
import {RatingApi} from "../../../shared/entities/rating/api/rating_api.ts";
import type {RatingChangeFormType} from "../../../shared/entities/rating/types/rating_types.ts";

export const UseChangeRating = () => {
    const queryclient = useQueryClient();
    
    return useMutation({
        mutationFn: (form: RatingChangeFormType) => RatingApi.changeRating(form),
        onSuccess: async()  => {
            await queryclient.invalidateQueries({queryKey: ["users_by_class"]});
            await queryclient.invalidateQueries({queryKey: ["user_by_id"]})
        }
    })
}