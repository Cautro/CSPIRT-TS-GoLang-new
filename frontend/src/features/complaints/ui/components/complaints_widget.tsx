import {useEffect} from "react";
import {useAuthStore} from "../../../auth/store/auth_store.ts";
import {useComplaintsStore} from "../../store/complaints_store.ts";
import {ComplaintCard} from "../../../../shared/ui/cards/complaint_card.tsx";

type props = {
    id: number,
    name: string,
}

export function ComplaintsWidget({id, name}: props) {
    const complaints = useComplaintsStore((state) => state.complaints)
    const getComplaints = useComplaintsStore((state) => state.getComplaints)
    const deleteComplaint = useComplaintsStore((state) => state.deleteComplaint);
    const role = useAuthStore((state) => state.user?.User.Role);
    const status = useComplaintsStore((state) => state.status);
    const error = useComplaintsStore((state) => state.error);

    const isLoading = status === "loading";

    useEffect(() => {
        void getComplaints(id);
    }, [getComplaints, id])

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
                <div className="alert alert--danger mb-4">{error}</div>
            )}

            {(complaints !== null && complaints.length > 0) ? (
                <div className={"class-list"}>
                    {complaints.map((item) => (
                        <ComplaintCard item={item} key={item.ID} onDelete={() => {
                            deleteComplaint(item.ID);
                            getComplaints(id);
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