import { z } from "zod";
import { ApiClient } from "../../../core/api/api_client.ts";
import { usersSchema, type UserType } from "../../../shared/entities/user/user_types.ts";
import {
    LOGIN_REGEX,
    SECURITY_LIMITS,
    normalizeText,
} from "../../../core/security/security_limits.ts";
import {noteSchema, type NoteType} from "../../../shared/entities/notes/notes_types.ts";

const errorResponseSchema = z.object({
    error: z.string().optional(),
    message: z.string().optional(),
}).passthrough();

const ratingResponseSchema = z.object({
    new_rating: z.number().int(),
    message: z.string().max(300),
});

const notesResponseSchema = z.object({
    All_notes: z.array(noteSchema),
});

export const ratingChangeSchema = z.object({
    rating: z
        .number()
        .int()
        .min(SECURITY_LIMITS.ratingDeltaMin)
        .max(SECURITY_LIMITS.ratingDeltaMax),
    target_login: z
        .string()
        .min(SECURITY_LIMITS.loginMin)
        .max(SECURITY_LIMITS.loginMax)
        .regex(LOGIN_REGEX),
    reason: z
        .string()
        .transform(normalizeText)
        .pipe(
            z.string()
                .min(SECURITY_LIMITS.ratingReasonMin)
                .max(SECURITY_LIMITS.ratingReasonMax),
        ),
});

export type RatingChangeDTO = z.infer<typeof ratingChangeSchema>;
export type RatingResponse = z.infer<typeof ratingResponseSchema>;

const client = new ApiClient();

function getApiError(data: unknown): string {
    const parsed = errorResponseSchema.safeParse(data);

    if (!parsed.success) {
        return "Ошибка сервера";
    }

    return parsed.data.error || parsed.data.message || "Ошибка сервера";
}

export const dashboardApi = {
    async getNotes(): Promise<NoteType[]> {
      const response = await client.get<unknown>("/api/notes", true);

      if (!response.checkStatus()) {
          throw new Error("Ошибка при получении списка заметок");
      }
      
      const parsed = notesResponseSchema.safeParse(response.data);
      if (!parsed.success) {
          throw new Error("Некорректный формат заметок");
      }
      
      return parsed.data.All_notes;
    },
    
    async getUsers(): Promise<UserType[]> {
        const response = await client.get<unknown>("/api/users", true);

        if (!response.checkStatus()) {
            throw new Error(getApiError(response.data));
        }

        const parsed = usersSchema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный формат пользователей");
        }

        return parsed.data;
    },

    async changeRating(data: RatingChangeDTO): Promise<RatingResponse> {
        const parsedDto = ratingChangeSchema.safeParse(data);

        if (!parsedDto.success) {
            throw new Error("Некорректные данные изменения рейтинга");
        }

        const response = await client.patch<unknown>(
            "/api/rating/update",
            parsedDto.data,
            true,
        );

        if (!response.checkStatus()) {
            throw new Error(getApiError(response.data));
        }

        const parsedResponse = ratingResponseSchema.safeParse(response.data);

        if (!parsedResponse.success) {
            throw new Error("Некорректный ответ сервера");
        }

        return parsedResponse.data;
    },
};