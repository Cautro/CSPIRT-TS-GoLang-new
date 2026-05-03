import type {EventType} from "../entities/events/types/events_types.ts";
import {format} from "date-fns";
import {ru} from "date-fns/locale";
import {useNavigate} from "react-router-dom";

interface Props {
    item: EventType
}

export function EventCard({ item }: Props) {
    const navigate = useNavigate();

    const date = format(new Date(item.StartedAt), "d MMMM yyyy, HH:mm", {
        locale: ru,
    });
    
    return (
        <button
            className="class-flat-card"
            type="button"
            onClick={() => {
                navigate("/event", {
                    state: {
                        eventId: item.ID
                    }
                })
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

                {item.Status === "completed" && (<div className="class-flat-card__metric">
                    <span className="class-flat-card__metric-label">Статус</span>
                    <span className="class-flat-card__metric-value">Завершено</span>
                </div>)}
                
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