import {useAuthStore} from "../../../auth/store/auth_store.ts";
import {ComplaintCard} from "../../../../shared/ui/cards/complaint_card.tsx";
import {useComplaints} from "../../hooks/use_complaints.ts";
import type {ComplaintType} from "../../../../shared/entities/complaints/types/complaints_types.ts";
import {useDeleteComplaint} from "../../hooks/use_delete_complaint.ts";
import {useEffect} from "react";

type props = {
    id: number,
    name: string,
}

export function ComplaintsWidget({id, name}: props) {
    const {mutate} = useDeleteComplaint();
    const role = useAuthStore((state) => state.user?.User.Role);
    const {data, isLoading, error} = useComplaints(id);
    const complaints = (data as ComplaintType[]) || [];
    
    useEffect(() => {
        console.log(complaints);
    }, [complaints]);
    
    return (
        <>
            {isLoading && (
                <div className="grid grid--3">
                    <div className="skeleton" style={{ height: 160 }} />
                    <div className="skeleton" style={{ height: 160 }} />
                    <div className="skeleton" style={{ height: 160 }} />
                </div>
            )}

            {error && !isLoading && (
                <div className="alert alert--danger mb-4">{error.message}</div>
            )}

            {(complaints !== null && complaints.length > 0) ? (
                <div className={"class-list"}>
                    {complaints.map((item) => (
                        <ComplaintCard item={item} key={item.ID} onDelete={() => {
                            mutate({id: item.ID})
                        }} role={role ?? "User"} />
                    ))}
                </div>
            ) :(
                !isLoading && <div className="empty-state">
                    <h2 className="empty-state__title">Жалобы не найдены</h2>
                    <p className="empty-state__text">
                        Не удалось найти жалобы по {name} классу
                    </p>
                </div>
            )}
        </>
    );
} 