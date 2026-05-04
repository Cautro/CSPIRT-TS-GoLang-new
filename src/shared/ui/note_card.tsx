import { format } from "date-fns";
import { ru } from "date-fns/locale";
import type { NoteType } from "../entities/notes/types/notes_types.ts";
import type {UserRole} from "../entities/user/types/user_types.ts";

interface Props {
    item: NoteType;
    onDelete?: () => void;
    role: UserRole;
}

export function NoteCard({ item, onDelete, role }: Props) {
    const date = format(new Date(item.CreatedAt), "d MMMM yyyy, HH:mm", {
        locale: ru,
    });

    const canDelete =
        Boolean(onDelete) && (role === "Owner" || role === "Admin");

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
                        onClick={onDelete}
                    >
                        Удалить
                    </button>
                )}
            </div>
        </div>
    );
}