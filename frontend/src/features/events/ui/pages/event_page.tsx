import {useEffect, useState} from "react";
import {useNavigate, useParams} from "react-router-dom";
import { format, isValid, parse } from "date-fns";
import { ru } from "date-fns/locale";

import { useEventStore } from "../../store/event_store.ts";
import { ClassCard } from "../../../../shared/ui/cards/class_card.tsx";
import {useAuthStore} from "../../../auth/store/auth_store.ts";
import {ConfirmModal} from "../../../../shared/ui/modals/confirm_modal.tsx";
import {useClassStore} from "../../../class/store/class_store.ts";
import {BurgerDrawerMenu, type BurgerDrawerMenuItem} from "../../../../shared/ui/other/burger_menu.tsx";

function getStatusLabel(status: string): string {
    if (!status.trim()) {
        return "Статус не указан";
    }

    const statusMap: Record<string, string> = {
        active: "Активно",
        completed: "Завершено",
        scheduled: "Запланировано",
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

    const { id } = useParams<{id: string}>();

    const classes = useClassStore((state) => state.classes);
    const event = useEventStore((state) => state.event);
    const getClasses = useClassStore((state) => state.getClasses);
    const getEvent = useEventStore((state) => state.getEventById);
    const completeEvent = useEventStore((state) => state.completeEvent);
    const deleteEvent = useEventStore((state) => state.deleteEvent);
    const error = useEventStore((state) => state.error);
    const status = useEventStore((state) => state.status);
    const role = useAuthStore((state) => state.user?.User.Role);

    const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
    const [isCompleteConfirmOpen, setIsCompleteConfirmOpen] = useState(false);
    
    const isLoading = status === "loading";
    
    const menuItems: BurgerDrawerMenuItem[] = [
        {
          label: "Завершить мероприятие",
          primary: true,
          hidden: (role !== "Owner"),
          onClick: () => setIsCompleteConfirmOpen(true),  
        },
        {
            label: "Удалить мероприятие",
            danger: true,
            hidden: (role !== "Owner"),
            onClick: () => setIsDeleteConfirmOpen(true),
        },
        {
            label: "Назад",
            onClick: () => navigate(-1)
        }
    ]

    useEffect(() => {
        if (!id) {
            return;
        }

        void getEvent(Number(id));
        void getClasses();
    }, [id, getClasses, getEvent]);

    if (!id) {
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
                            Не удалось найти мероприятие с ID {id}.
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
    const eventClasses = classes?.filter((item) => eventClassIds.has(item.Id));


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

                            <BurgerDrawerMenu items={menuItems} title={"Меню"} />
                            
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

                    {!isLoading && !error && eventClasses?.length > 0 && role === "Owner" && (
                        <div className="class-list">
                            {eventClasses?.map((item) => (
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

                    {!isLoading && !error && eventClasses?.length === 0 && (
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
                    <ConfirmModal
                        title={"Удалить мероприятие?"}
                        content={`Это действие удалит мероприятие "${event.Title}". Отменить удаление будет нельзя.`}
                        onConfirm={async () => {
                            await deleteEvent(event.ID);
                            setIsDeleteConfirmOpen(false);
                            navigate(-1);
                        }}
                        isOpen={isDeleteConfirmOpen}
                        onClose={() => setIsDeleteConfirmOpen(false)}
                        buttonContent={"Удалить"}
                        isDanger={true}
                    />
                )}

                
                {isCompleteConfirmOpen && (
                    <ConfirmModal 
                        title={"Завершить мероприятие?"}
                        content={`После завершения участникам мероприятия "${event.Title}" будет начислена награда: +${event.RatingReward} рейтинга.`}
                        onConfirm={async () => {
                            await completeEvent(event);
                            setIsCompleteConfirmOpen(false);
                            navigate(-1);
                        }}
                        isOpen={isCompleteConfirmOpen}
                        onClose={() => setIsCompleteConfirmOpen(false)}
                        buttonContent={"Завершить"}
                    />
                )}
                
            </section>
        </main>
    );
}