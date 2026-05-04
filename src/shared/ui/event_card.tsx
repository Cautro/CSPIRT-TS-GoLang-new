import type {EventType} from "../entities/events/types/events_types.ts";
import {format} from "date-fns";
import {ru} from "date-fns/locale";

interface Props {
    item: EventType
    onClick?: () => void
}

export function EventCard({ item, onClick }: Props) {
    function getStatusLabel(status: string): string {
        if (!status.trim()) {
            return "Статус не указан";
        }

        const statusMap: Record<string, string> = {
            active: "Активно",
            completed: "Завершено",
            planned: "Запланировано",
        };

        return statusMap[status] ?? status;
    }
    
    const status = getStatusLabel(item.Status);

    const date = format(new Date(item.StartedAt), "d MMMM yyyy, HH:mm", {
        locale: ru,
    });
    
    return (
        <button
            className="class-flat-card"
            type="button"
            onClick={() => {
                if (onClick) {
                    void onClick();
                }
            }}
        >
            <div className="class-flat-card__main">
                <div className="class-flat-card__info">
                    <p className="class-flat-card__subtitle">
                        {item.Title}
                    </p>
                </div>
            </div>

            <div className="class-flat-card__meta">

                <div className="class-flat-card__metric">
                    <span className="class-flat-card__metric-label">Статус</span>
                    <span className="class-flat-card__metric-value">{status}</span>
                </div>
                
                <div className="class-flat-card__metric">
                    <span className="class-flat-card__metric-label">Количество участников</span>
                    <span className="class-flat-card__metric-value">{item.Players.length}</span>
                </div>

                <div className="class-flat-card__metric">
                    <span className="class-flat-card__metric-label">Награда за участие</span>
                    <span className="class-flat-card__metric-value">{item.RatingReward}</span>
                </div>

                <div className="class-flat-card__metric">
                    <span className="class-flat-card__metric-label">Начало мероприятия</span>
                    <span className="class-flat-card__metric-value">{date}</span>
                </div>

                <span className="class-flat-card__arrow">→</span>
            </div>
        </button>
    );
}