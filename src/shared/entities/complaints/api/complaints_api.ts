import {z} from "zod";
import {ApiClient} from "../../../../core/api/api_client.ts";
import {complaintSchema, type ComplaintType} from "../types/complaints_types.ts";

export const complaintAddDto = z.object({
    AuthorID: z.number().int().nonnegative(),
    CreatedAt: z.string(),
    TargetID: z.number().int().nonnegative(),
    Content: z.string().max(500),
    AuthorName: z.string(),
    TargetName: z.string(),
});

export type complaintAddType = z.infer<typeof complaintAddDto>

const complaintsResponseShema = z.object({
    Complaints: z.array(complaintSchema),
});

const complaintAddResponseSchema = z.object({
    messsage: z.string(),
});

const complaintDeleteResponseSchema = z.object({
    messsage: z.string(),
});

const client = new ApiClient();

export const ComplaintsApi = {
    async getComplaints(id: string): Promise<ComplaintType[]> {
        const response = await client.get(`/api/complaints?class=${id}`, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении списка жалоб");
        }

        const parsed = complaintsResponseShema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный формат жалоб");
        }

        return parsed.data.Complaints;
    },

    async addComplaint(dto: complaintAddType): Promise<boolean> {
        const response = await client.patch("/api/complaint/add", dto, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при добавлении жалобы");
        }

        const parsed = complaintAddResponseSchema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }

        return true;
    },

    async deleteComplaint(id: string): Promise<boolean> {
        const response = await client.delete(`/api/complaint/delete/${id}`, {}, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при удалении жалобы");
        }

        const parsed = complaintDeleteResponseSchema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный ответ сервера");
        }

        return true;
    }
}