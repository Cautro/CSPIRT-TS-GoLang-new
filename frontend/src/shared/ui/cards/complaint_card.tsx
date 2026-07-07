import { format } from 'date-fns';
import { ru } from 'date-fns/locale';
import type {ComplaintType} from "../../entities/complaints/types/complaints_types.ts";
import type {UserRole} from "../../entities/user/types/user_types.ts";
import {useState} from "react";
import {ConfirmModal} from "../modals/confirm_modal.tsx";

interface Props {
    item: ComplaintType;
    onDelete?: () => void;
    role?: UserRole;
}

export function ComplaintCard({ item, onDelete, role }: Props) {
    const date = format(new Date(item.CreatedAt), 'd MMMM yyyy, HH:mm', { locale: ru });
    const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
    const normalizedRole = role?.toLowerCase();

    const canDelete =
        Boolean(onDelete) && (normalizedRole === "owner" || normalizedRole === "admin");

    return (
        <div className="class-flat-card">
            <div className="class-flat-card__main">
                <div className="class-flat-card__icon">
                    {item.ID}
                </div>

                <div className="class-flat-card__info">
                    <h2 className="class-flat-card__subtitle">
                        От {item.AuthorName}
                    </h2>

                    <p className="class-flat-card__subtitle">
                        На {item.TargetName}
                    </p>
                </div>

                <div className="class-flat-card__info">
                    <h2 className="class-flat-card__subtitle">
                        {item.Content}
                    </h2>
                </div>
            </div>

            <div className="class-flat-card__meta">
                <div className="class-flat-card__metric">
                    <span className="class-flat-card__metric-label">
                        Создана
                    </span>
                    <span className="class-flat-card__metric-value">
                        {date}
                    </span>
                </div>

                {canDelete && (
                    <button
                        className="btn btn--danger"
                        onClick={() => setIsDeleteModalOpen(true)}
                    >
                        Удалить
                    </button>
                )}
            </div>
            
            <ConfirmModal
                title={"Удаление жалобы"}
                content={"Вы уверены что хотите удалить жалобу? Это действие нельзя отменить."}
                onConfirm={async () => {if (onDelete) onDelete()}}
                onClose={() => setIsDeleteModalOpen(false)}
                isOpen={isDeleteModalOpen}
                buttonContent={"Удалить"}
                isDanger={true}
            />
        </div>
    );
}