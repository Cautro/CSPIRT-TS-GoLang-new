import {useNavigate} from "react-router-dom";
import {EventCard} from "../../../../shared/ui/cards/event_card.tsx";
import {UseEvents} from "../../hooks/use_events.ts";

export function EventsWidget() {
    const navigate = useNavigate();
    const {error: error, data: events, isLoading: isLoading} = UseEvents();
    
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

            {events && !isLoading && !error && events.length > 0 ? (
                <div className="class-list">
                    {events.map((item) => (
                        <EventCard key={item.ID} item={item} onClick={() => navigate(`/event/${item.ID}`)} />
                    ))}
                </div>
            ) : (
                !isLoading && !error && (events?.length === 0 || !events) &&  <div className="empty-state">
                    <h2 className="empty-state__title">Мероприятия не найдены</h2>
                    <p className="empty-state__text">
                        Не удалось найти доступные мероприятия
                    </p>
                </div>
            )}
        </>
    );
} 