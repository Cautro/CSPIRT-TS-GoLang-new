import {z} from "zod";
import {ApiClient} from "../../../../core/api/api_client.ts";

export const ratingChangeDTO = z.object({
    rating: z.number().int().max(5000).min(-5000),
    target_login: z.string(),
    reason: z.string(),
});

const ratingChangeResponse = z.object({
    message: z.string(),
    new_rating: z.number().int().nonnegative(),
    target: z.string(),
});

export type ratingChangeType = z.infer<typeof ratingChangeDTO>

const client = new ApiClient();

export const RatingApi = {
    async changeRating(dto: ratingChangeType): Promise<string> {
        const response = await client.patch("/api/rating/update", dto, true);
        
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