import {useState} from "react";
import {useNavigate, useParams} from "react-router-dom";
import { format, isValid, parse } from "date-fns";
import { ru } from "date-fns/locale";
import { ClassCard } from "../../../../shared/ui/cards/class_card.tsx";
import {useAuthStore} from "../../../auth/store/auth_store.ts";
import {ConfirmModal} from "../../../../shared/ui/modals/confirm_modal.tsx";
import {type BurgerDrawerMenuItem} from "../../../../shared/ui/other/burger_menu.tsx";
import {PageHeader} from "../../../../shared/ui/other/page_header.tsx";
import {useClasses} from "../../../class/hooks/use_classes.ts";
import {UseEventById} from "../../hooks/use_event_by_id.ts";
import {useCompleteEvent} from "../../hooks/use_complete_event.ts";
import {useDeleteEvent} from "../../hooks/use_delete_event.ts";
import {UseRewardParams} from "../../hooks/use_reward_params.ts";
import {AddParamModal} from "../components/add_param_modal.tsx";

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
    
    const classes = useClasses().data;
    const getEvent = UseEventById(Number(id))
    const event = getEvent.data
    const completeEvent = useCompleteEvent()
    const deleteEvent = useDeleteEvent()
    const error = getEvent.error?.message || completeEvent.error?.message || deleteEvent.error?.message;
    const user = useAuthStore((state) => state.user?.User);
    const normalizedRole = user?.Role.toLowerCase();
    
    const getRewardParams = UseRewardParams(Number(id));
    const rewardParams = getRewardParams.data;

    const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
    const [isCompleteConfirmOpen, setIsCompleteConfirmOpen] = useState(false);
    const [isAddRewardModalOpen, setIsAddRewardModalOpen] = useState(false);
    
    const isLoading = getEvent.isLoading || getRewardParams.isLoading;
    
    const menuItems: BurgerDrawerMenuItem[] = [
        {
          label: "Завершить мероприятие",
          hidden: (normalizedRole !== "owner"),
          onClick: () => setIsCompleteConfirmOpen(true),  
        },
        {
            label: "Изменить награду для класса",
            hidden: (normalizedRole !== "owner"),
            onClick: () => setIsAddRewardModalOpen(true),
        },
        {
            label: "Удалить мероприятие",
            hidden: (normalizedRole !== "owner"),
            onClick: () => setIsDeleteConfirmOpen(true),
        },
    ]

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

    const eventClasses = (classes ?? []).filter((item) => {
        return eventClassIds.has(item.Id);
    });
    
    const statusLabel = getStatusLabel(event.Status);
    const date = formatStartedAt(event.StartedAt);

    return (
        <main className="main">
            <section className="page">
                <div className="event-page">
                    
                    <PageHeader
                        eyebrow={`Мероприятие #${event.ID}`}
                        title={event.Title}
                        description={event.Description} 
                        menuTitle={`Меню мероприятия`}
                        menuItems={menuItems}
                        hasBackButton={true}
                    />

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
                    
                    {!isLoading && !error && eventClasses?.length > 0 && (
                        <div className="class-list">
                            {eventClasses?.map((item) => {
                                let params = null;
                                rewardParams?.map((param) => {
                                    if (param.ClassID === item.Id) {
                                        params = param.Reason + ` Награда: ${param.ExtraRatingReward}`
                                    }
                                })
                                
                                if (params) {
                                    return <ClassCard
                                        key={item.Id}
                                        item={item}
                                        param={params}
                                        onClick={() => {
                                            if ((normalizedRole === "owner") || (user?.Id === item.Teacher?.Id)) {
                                                navigate(
                                                    `/events/${event.ID}/classes/${item.Id}/players/add`,
                                                    {
                                                        state: {
                                                            event,
                                                            classItem: item,
                                                        },
                                                    }
                                                );
                                            }
                                        }}
                                    />
                                } else {
                                    return <ClassCard
                                        key={item.Id}
                                        item={item}
                                        onClick={() => {
                                            if ((normalizedRole === "owner") || (user?.Id === item.Teacher?.Id)) {
                                                navigate(
                                                    `/events/${event.ID}/classes/${item.Id}/players/add`,
                                                    {
                                                        state: {
                                                            event,
                                                            classItem: item,
                                                        },
                                                    }
                                                );
                                            }
                                        }}
                                    />
                                }
                                
                            })}
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

                {isAddRewardModalOpen && (
                    <AddParamModal 
                        isOpen={isAddRewardModalOpen} 
                        onClose={() => setIsAddRewardModalOpen(false)} 
                        Classes={eventClasses} 
                        EventID={Number(id)}
                        OnAdd={() => setIsAddRewardModalOpen(false)}
                    />
                )}


                {isDeleteConfirmOpen && (
                    <ConfirmModal
                        title={"Удалить мероприятие?"}
                        content={`Это действие удалит мероприятие "${event.Title}". Отменить удаление будет нельзя.`}
                        onConfirm={async () => {
                            await deleteEvent.mutateAsync({id: event.ID});
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
                            await completeEvent.mutateAsync({item: event});
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