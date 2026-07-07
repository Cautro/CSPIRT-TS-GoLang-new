import {z} from "zod";
import {apiClient} from "../../../../core/api/client.ts";
import type {RatingChangeFormType} from "../types/rating_types.ts";

const ratingChangeResponse = z.object({
    message: z.string(),
    new_rating: z.number().int().nonnegative(),
    target: z.string(),
});

export const RatingApi = {
    async changeRating(dto: RatingChangeFormType): Promise<string> {
        const response = await apiClient.patch("/api/rating/update", dto, true);
        
        if (!response.checkStatus()) {
            throw new Error("Ошибка при изменении рейтинга");
        }
        
        const parsed = ratingChangeResponse.safeParse(response.data);
        
        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }
        
        return `Новый рейтинг для ${parsed.data.target} - ${parsed.data.new_rating}`;
    }
}