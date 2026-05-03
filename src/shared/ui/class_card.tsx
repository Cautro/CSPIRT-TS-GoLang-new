import type { ClassType } from "../entities/class/types/class_types.ts";

interface Props {
    item: ClassType;
    onClick?: () => void;
}

export function ClassCard({ item, onClick }: Props) {
    const teacherFullName = item.Teacher
        ? `${item.Teacher.Name} ${item.Teacher.LastName}`
        : "Не назначен";

    const membersCount = item.Members.length;

    return (
        <button
            className="class-flat-card"
            type="button"
            onClick={onClick}
        >
            <div className="class-flat-card__main">
                <div className="class-flat-card__icon">
                    {item.Name}
                </div>

                <div className="class-flat-card__info">
                    <h2 className="class-flat-card__title">{item.Name} Класс</h2>

                    <p className="class-flat-card__subtitle">
                        Классный руководитель: {teacherFullName}
                    </p>
                </div>
            </div>

            <div className="class-flat-card__meta">
                <div className="class-flat-card__metric">
                    <span className="class-flat-card__metric-label">Ученики</span>
                    <span className="class-flat-card__metric-value">{membersCount}</span>
                </div>

                <div className="class-flat-card__metric">
                    <span className="class-flat-card__metric-label">Рейтинг</span>
                    <span className="class-flat-card__metric-value">{item.TotalRating}</span>
                </div>

                <span className="class-flat-card__arrow">→</span>
            </div>
        </button>
    );
}