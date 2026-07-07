import {z} from "zod";
import {complaintSchema, type ComplaintType} from "../types/complaints_types.ts";
import {apiClient} from "../../../../core/api/client.ts";


export type complaintAddFormType = { 
    AuthorID: number;
    CreatedAt: string;
    TargetID: number;
    Content: string;
    AuthorName: string;
    TargetName: string;
}

const complaintsResponseShema = z.object({
    Complaints: z.array(complaintSchema),
});

export const ComplaintsApi = {
    async getComplaints(id: number): Promise<ComplaintType[]> {
        const response = await apiClient.get(`/api/complaints?class=${id}`, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при получении списка жалоб");
        }

        const parsed = complaintsResponseShema.safeParse(response.data);

        if (!parsed.success) {
            throw new Error("Некорректный формат жалоб");
        }

        return parsed.data.Complaints;
    },

    async addComplaint(dto: complaintAddFormType): Promise<boolean> {
        console.log(dto.TargetID, dto.Content);
        
        const response = await apiClient.patch("/api/complaint/add", {
            TargetID: dto.TargetID,
            Content: dto.Content,
        }, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при добавлении жалобы");
        }

        return true;
    },

    async deleteComplaint(id: number): Promise<boolean> {
        const response = await apiClient.delete(`/api/complaint/delete/${id}`, {}, true);

        if (!response.checkStatus()) {
            throw new Error("Ошибка при удалении жалобы");
        }

        return true;
    }
}