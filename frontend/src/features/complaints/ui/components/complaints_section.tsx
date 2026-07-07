import {ComplaintCard} from "../../../../shared/ui/cards/complaint_card.tsx";
import type {ComplaintType} from "../../../../shared/entities/complaints/types/complaints_types.ts";
import {useState} from "react";
import {useDeleteComplaint} from "../../hooks/use_delete_complaint.ts";
import {useUser} from "../../../users/hooks/use_user.ts";
import type {UserType} from "../../../../shared/entities/user/types/user_types.ts";
import {ComplaintAddUsecase} from "../../models/complaint_add_usecase.ts";
import {useAddComplaint} from "../../hooks/use_add_complaint.ts";

export const ComplaintsSection = ({complaints, isYou, user, currentUser, setFormError = () => {}, isProfile = false}:  {isProfile?: boolean, setFormError?: (value: string | null) => void, currentUser: UserType, complaints: ComplaintType[], isYou: boolean, user: UserType})=> {
    const [complaintText, setComplaintText] = useState<string>("");
    const deleteComplaint = useDeleteComplaint();
    const getUser = useUser(user.Id);
    const addComplaint = useAddComplaint();

    async function handleComplaintAdd() {
        setFormError(null);

        if (!user || !currentUser) {
            setFormError("Не удалось определить пользователя");
            return;
        }

        const form = {
            user: user,
            current_user: currentUser,
            content: complaintText,
        };

        try {
            const dto = ComplaintAddUsecase(form);
            await addComplaint.mutateAsync(dto);
            setComplaintText("");
            await getUser.refetch();
        } catch (e) {
            setFormError(e instanceof Error ? e.message : "Неизвестная ошибка");
        }
    }
    
    return (
        <section className="card card--padded user-section-card">
            <div className="section-head section-head--row">
                <div>
                    <h2 className="section-title">Жалобы</h2>
                    <p className="section-description">
                        Жалобы, связанные с текущим пользователем.
                    </p>
                </div>

                <span
                    className={
                        complaints.length > 0
                            ? "badge badge--danger"
                            : "badge badge--neutral"
                    }
                >
                    {complaints.length}
                </span>
            </div>

            {(isProfile === false && !isYou) && (<div className="user-action-form">
                <div className="field">
                    <label className="field__label" htmlFor="complaintText">
                        Новая жалоба
                    </label>

                    <textarea
                        id="complaintText"
                        className="textarea"
                        placeholder="Введите текст жалобы на пользователя..."
                        minLength={6}
                        maxLength={500}
                        value={complaintText}
                        onChange={(event) => setComplaintText(event.target.value)}
                    />
                </div>

                <div className="user-action-form__footer">
                    <button
                        className="btn btn--danger"
                        type="button"
                        disabled={!complaintText.trim()}
                        onClick={() => void handleComplaintAdd()}
                    >
                        Отправить жалобу
                    </button>
                </div>
            </div>)}

            {complaints.length > 0 ? (
                <div className="feed">
                    {isProfile === false ? complaints.map((item) => (
                        <ComplaintCard
                            key={item.ID}
                            item={item}
                            onDelete={async () => {
                                await deleteComplaint.mutateAsync({id: item.ID});
                                await getUser.refetch()
                            }}
                            role={currentUser.Role ?? "User"}
                        />
                    )) : complaints.map((item) => (
                        <ComplaintCard
                            key={item.ID}
                            item={item}
                            role={currentUser.Role ?? "User"}
                        />
                    ))}
                </div>
            ) : (
                <div className="empty-inline">Жалоб нет</div>
            )}
        </section>
    )
}