import {useEffect, useState} from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { format, isValid, parse } from "date-fns";
import { ru } from "date-fns/locale";

import { useEventStore } from "../../store/event_store.ts";
import { ClassCard } from "../../../../shared/ui/class_card.tsx";
import {useAuthStore} from "../../../auth/store/auth_store.ts";

interface EventPageState {
    eventId?: number;
}

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

function formatStartedAt(value: string): string {
    const parsedDate = parse(value, "yyyy-MM-dd HH:mm", new Date());

    if (isValid(parsedDate)) {
        return format(parsedDate, "d MMMM yyyy, HH:mm", { locale: ru });
    }

    const fallbackDate = new Date(value);

    if (isValid(fallbackDate)) {
        return format(fallbackDate, "d MMMM yyyy, HH:mm", { locale: ru });
    }

    return "Дата не указана";
}

export function EventPage() {
    const navigate = useNavigate();
    const location = useLocation();

    const { eventId } = (location.state ?? {}) as EventPageState;

    const classes = useEventStore((state) => state.classes);
    const event = useEventStore((state) => state.event);
    const getClasses = useEventStore((state) => state.getClasses);
    const getEvent = useEventStore((state) => state.getEventById);
    const completeEvent = useEventStore((state) => state.completeEvent);
    const deleteEvent = useEventStore((state) => state.deleteEvent);
    const error = useEventStore((state) => state.error);
    const status = useEventStore((state) => state.status);
    const role = useAuthStore((state) => state.user?.User.Role);

    const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
    const [isCompleteConfirmOpen, setIsCompleteConfirmOpen] = useState(false);
    
    const isLoading = status === "loading";

    useEffect(() => {
        if (!eventId) {
            return;
        }

        void getEvent(eventId);
        void getClasses();
    }, [eventId, getClasses, getEvent]);

    if (!eventId) {
        return (
            <main className="main">
                <section className="page">
                    <div className="empty-state">
                        <h2 className="empty-state__title">
                            ID мероприятия не найден
                        </h2>

                        <p className="empty-state__text">
                            Страница открыта без данных о мероприятии.
                        </p>

                        <button
                            className="btn btn--primary"
                            type="button"
                            onClick={() => navigate(-1)}
                        >
                            Вернуться назад
                        </button>
                    </div>
                </section>
            </main>
        );
    }

    if (isLoading && !event) {
        return (
            <main className="main">
                <section className="page">
                    <div className="class-list">
                        <div className="skeleton" style={{ height: 120 }} />
                        <div className="skeleton" style={{ height: 88 }} />
                        <div className="skeleton" style={{ height: 88 }} />
                    </div>
                </section>
            </main>
        );
    }

    if (error && !event) {
        return (
            <main className="main">
                <section className="page">
                    <div className="alert alert--danger">
                        {error}
                    </div>

                    <button
                        className="btn btn--primary"
                        type="button"
                        onClick={() => navigate(-1)}
                    >
                        Вернуться назад
                    </button>
                </section>
            </main>
        );
    }

    if (!event) {
        return (
            <main className="main">
                <section className="page">
                    <div className="empty-state">
                        <h2 className="empty-state__title">
                            Мероприятие не найдено
                        </h2>

                        <p className="empty-state__text">
                            Не удалось найти мероприятие с ID {eventId}.
                        </p>

                        <button
                            className="btn btn--primary"
                            type="button"
                            onClick={() => navigate(-1)}
                        >
                            Вернуться назад
                        </button>
                    </div>
                </section>
            </main>
        );
    }

    const eventClassIds = new Set(event.Classes ?? []);
    const eventClasses = classes.filter((item) => eventClassIds.has(item.Id));


    const statusLabel = getStatusLabel(event.Status);
    const date = formatStartedAt(event.StartedAt);

    return (
        <main className="main">
            <section className="page">
                <div className="event-page">
                    <div className="event-page__header">
                        <div>
                            <p className="event-page__eyebrow">
                                Мероприятие #{event.ID}
                            </p>

                            <h1 className="event-page__title">
                                {event.Title}
                            </h1>

                            <p className="event-page__description">
                                {event.Description || "Описание не указано"}
                            </p>
                        </div>
                        
                        <div className="btn-group">

                            {role === "Owner" && (
                                <div className="btn-group">
                                    <button
                                        className="btn btn--danger"
                                        type="button"
                                        onClick={() => setIsDeleteConfirmOpen(true)}
                                        disabled={isLoading}
                                    >
                                        Удалить мероприятие
                                    </button>

                                    {event.Status !== "completed" && (
                                        <button
                                            className="btn btn--primary"
                                            type="button"
                                            onClick={() => setIsCompleteConfirmOpen(true)}
                                            disabled={isLoading}
                                        >
                                            Завершить мероприятие
                                    </button>
                                        )}
                                </div>
                        )}
    
                            <button
                                className="btn btn--secondary"
                                type="button"
                                onClick={() => navigate(-1)}
                            >
                                Назад
                            </button>
                        </div>

                    </div>

                    <div className="event-page__grid">
                        <article className="event-info-card">
                            <p className="event-info-card__label">
                                Статус
                            </p>

                            <p className="event-info-card__value">
                                {statusLabel}
                            </p>
                        </article>

                        <article className="event-info-card">
                            <p className="event-info-card__label">
                                Награда за участие
                            </p>

                            <p className="event-info-card__value">
                                +{event.RatingReward} рейтинга
                            </p>
                        </article>

                        <article className="event-info-card">
                            <p className="event-info-card__label">
                                Дата начала
                            </p>

                            <p className="event-info-card__value">
                                {date}
                            </p>
                        </article>

                        <article className="event-info-card">
                            <p className="event-info-card__label">
                                Количество участников
                            </p>

                            <p className="event-info-card__value">
                                {event.Players?.length ?? 0}
                            </p>
                        </article>
                    </div>

                    {error && (
                        <div className="alert alert--danger">
                            {error}
                        </div>
                    )}

                    {!isLoading && !error && eventClasses.length > 0 && role === "Owner" && (
                        <div className="class-list">
                            {eventClasses.map((item) => (
                                <ClassCard
                                    key={item.Id}
                                    item={item}
                                    onClick={() => {
                                        navigate(
                                            `/events/${event.ID}/classes/${item.Id}/players/add`,
                                            {
                                                state: {
                                                    event,
                                                    classItem: item,
                                                },
                                            }
                                        );
                                    }}
                                />
                            ))}
                        </div>
                    )}

                    {!isLoading && !error && eventClasses.length === 0 && (
                        <div className="empty-state">
                            <h2 className="empty-state__title">
                                Классы не найдены
                            </h2>

                            <p className="empty-state__text">
                                Для мероприятия пока не указаны классы.
                            </p>
                        </div>
                    )}
                </div>

                {isDeleteConfirmOpen && (
                    <div className="modal-backdrop" onMouseDown={() => setIsDeleteConfirmOpen(false)}>
                        <section
                            className="modal modal--confirm"
                            role="dialog"
                            aria-modal="true"
                            aria-labelledby="delete-event-title"
                            onMouseDown={(event) => event.stopPropagation()}
                        >
                            <div className="modal__header">
                                <div>
                                    <h2 className="modal__title" id="delete-event-title">
                                        Удалить мероприятие?
                                    </h2>

                                    <p className="modal__description">
                                        Это действие удалит мероприятие «{event.Title}». Отменить удаление будет нельзя.
                                    </p>
                                </div>

                                <button
                                    className="modal__close"
                                    type="button"
                                    onClick={() => setIsDeleteConfirmOpen(false)}
                                    aria-label="Закрыть окно подтверждения"
                                    disabled={isLoading}
                                >
                                    ×
                                </button>
                            </div>

                            <div className="modal__footer">
                                <button
                                    className="btn btn--secondary"
                                    type="button"
                                    onClick={() => setIsDeleteConfirmOpen(false)}
                                    disabled={isLoading}
                                >
                                    Отмена
                                </button>

                                <button
                                    className="btn btn--danger"
                                    type="button"
                                    disabled={isLoading}
                                    onClick={async () => {
                                        await deleteEvent(event.ID);
                                        setIsDeleteConfirmOpen(false);
                                        navigate(-1);
                                    }}
                                >
                                    {isLoading ? "Удаление..." : "Удалить"}
                                </button>
                            </div>
                        </section>
                    </div>
                )}

                {isCompleteConfirmOpen && (
                    <div className="modal-backdrop" onMouseDown={() => setIsCompleteConfirmOpen(false)}>
                        <section
                            className="modal modal--confirm"
                            role="dialog"
                            aria-modal="true"
                            aria-labelledby="complete-event-title"
                            onMouseDown={(event) => event.stopPropagation()}
                        >
                            <div className="modal__header">
                                <div>
                                    <h2 className="modal__title" id="complete-event-title">
                                        Завершить мероприятие?
                                    </h2>

                                    <p className="modal__description">
                                        После завершения участникам мероприятия «{event.Title}» будет начислена награда:
                                        {" "}
                                        +{event.RatingReward} рейтинга.
                                    </p>
                                </div>

                                <button
                                    className="modal__close"
                                    type="button"
                                    onClick={() => setIsCompleteConfirmOpen(false)}
                                    aria-label="Закрыть окно подтверждения"
                                    disabled={isLoading}
                                >
                                    ×
                                </button>
                            </div>

                            <div className="modal__footer">
                                <button
                                    className="btn btn--secondary"
                                    type="button"
                                    onClick={() => setIsCompleteConfirmOpen(false)}
                                    disabled={isLoading}
                                >
                                    Отмена
                                </button>

                                <button
                                    className="btn btn--primary"
                                    type="button"
                                    disabled={isLoading}
                                    onClick={async () => {
                                        await completeEvent(event);
                                        setIsCompleteConfirmOpen(false);
                                        navigate(-1);
                                    }}
                                >
                                    {isLoading ? "Завершение..." : "Завершить"}
                                </button>
                            </div>
                        </section>
                    </div>
                )}
                
            </section>
        </main>
    );
}