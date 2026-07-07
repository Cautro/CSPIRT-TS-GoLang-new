import {useMutation, useQueryClient} from "@tanstack/react-query";
import {type complaintAddFormType, ComplaintsApi} from "../../../shared/entities/complaints/api/complaints_api.ts";

export const useAddComplaint = () => {
    const queryclient = useQueryClient();
    
    return useMutation({
        mutationFn: (form: complaintAddFormType) => ComplaintsApi.addComplaint(form),
        onSuccess: async () => await queryclient.invalidateQueries({queryKey: ["complaints"]}),
    })
}