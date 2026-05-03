import { useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import type { ClassType } from "../../../../shared/entities/class/types/class_types.ts";
import type { UserType } from "../../../../shared/entities/user/types/user_types.ts";
import { addEventPlayersSchema } from "../../../../shared/entities/events/api/events_api.ts";
import { useEventStore } from "../../store/event_store.ts";

interface EventData {
    ID: number;
    Title: string;
    Status: string;
    RatingReward: number;
    Description: string;
    CreatedAt: string;
    StartedAt: string;
    Players?: unknown[];
    Classes?: number[];
}

interface PageState {
    event?: EventData;
    classItem?: ClassType;
}

function getPlayerId(player: unknown): number | null {
    if (typeof player === "number") {
        return player;
    }

    if (typeof player === "object" && player !== null) {
        const data = player as { Id?: unknown; ID?: unknown; UserID?: unknown };

        if (typeof data.Id === "number") {
            return data.Id;
        }

        if (typeof data.ID === "number") {
            return data.ID;
        }

        if (typeof data.UserID === "number") {
            return data.UserID;
        }
    }

    return null;
}

function uniqueNumbers(values: number[]): number[] {
    return Array.from(new Set(values));
}

export function EventClassPlayersPage() {
    const navigate = useNavigate();
    const location = useLocation();

    const { event, classItem } = (location.state ?? {}) as PageState;

    const status = useEventStore((state) => state.status);
    const error = useEventStore((state) => state.error);
    const addPlayersToEvent = useEventStore((state) => state.addPlayersToEvent);
    const removePlayersFromEvent = useEventStore((state) => state.removePlayersFromEvent);

    const [selectedUserIds, setSelectedUserIds] = useState<number[]>([]);
    const [formError, setFormError] = useState<string | null>(null);

    const isLoading = status === "loading";

    const students = useMemo<UserType[]>(() => {
        if (!classItem) {
            return [];
        }

        return classItem.Members.filter((user) => {
            return user.Role === "User" || user.Role === "Helper";
        });
    }, [classItem]);

    const initialSelectedUserIds = useMemo<number[]>(() => {
        if (!event?.Players) {
            return [];
        }

        return uniqueNumbers(
            event.Players
                .map(getPlayerId)
                .filter((id): id is number => typeof id === "number")
        );
    }, [event]);

    useEffect(() => {
        setSelectedUserIds(initialSelectedUserIds);
    }, [initialSelectedUserIds]);

    if (!event || !classItem) {
        return (
            <main className="main">
                <section className="page">
                    <div className="empty-state">
                        <h2 className="empty-state__title">
                            Данные не найдены
                        </h2>

                        <p className="empty-state__text">
                            Не удалось получить мероприятие или класс.
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

    function toggleUser(userId: number) {
        setSelectedUserIds((prev) => {
            if (prev.includes(userId)) {
                return prev.filter((id) => id !== userId);
            }

            return [...prev, userId];
        });
    }

    function selectAll() {
        setSelectedUserIds(students.map((item) => item.Id));
    }

    function clearSelected() {
        setSelectedUserIds([]);
    }

    async function handleSaveChanges() {
        setFormError(null);

        const initialSet = new Set(initialSelectedUserIds);
        const selectedSet = new Set(selectedUserIds);

        const addedIds = selectedUserIds.filter((id) => !initialSet.has(id));
        const removedIds = initialSelectedUserIds.filter((id) => !selectedSet.has(id));

        if (addedIds.length === 0 && removedIds.length === 0) {
            setFormError("Изменений нет");
            return;
        }

        const addDto = {
            playerIds: addedIds,
        };

        const removeDto = {
            playerIds: removedIds,
        };

        if (addedIds.length > 0) {
            const parsedAdd = addEventPlayersSchema.safeParse(addDto);

            if (!parsedAdd.success) {
                console.log(parsedAdd.error.issues);
                setFormError("Некорректный список учеников для добавления");
                return;
            }

            await addPlayersToEvent(event.ID, parsedAdd.data);
        }

        if (removedIds.length > 0) {
            const parsedRemove = addEventPlayersSchema.safeParse(removeDto);

            if (!parsedRemove.success) {
                console.log(parsedRemove.error.issues);
                setFormError("Некорректный список учеников для удаления");
                return;
            }

            await removePlayersFromEvent(event.ID, parsedRemove.data);
        }

        navigate(-1);
    }

    return (
        <main className="main">
            <section className="page">
                <div className="event-players-page">
                    <div className="event-page__header">
                        <div>
                            <p className="event-page__eyebrow">
                                Мероприятие #{event.ID}
                            </p>

                            <h1 className="event-page__title">
                                Участники мероприятия
                            </h1>

                            <p className="event-page__description">
                                {event.Title} · {classItem.Name} класс
                            </p>
                        </div>

                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => navigate(-1)}
                            disabled={isLoading}
                        >
                            Назад
                        </button>
                    </div>

                    <div className="event-players-toolbar">
                        <div>
                            <h2 className="event-players-toolbar__title">
                                Список учеников
                            </h2>

                            <p className="event-players-toolbar__text">
                                Выбрано: {selectedUserIds.length} из {students.length}
                            </p>
                        </div>

                        <div className="btn-group">
                            <button
                                className="btn btn--secondary"
                                type="button"
                                onClick={selectAll}
                                disabled={isLoading || students.length === 0}
                            >
                                Выбрать всех
                            </button>

                            <button
                                className="btn btn--secondary"
                                type="button"
                                onClick={clearSelected}
                                disabled={isLoading || selectedUserIds.length === 0}
                            >
                                Снять выбор
                            </button>

                            <button
                                className="btn btn--primary"
                                type="button"
                                onClick={handleSaveChanges}
                                disabled={isLoading}
                            >
                                {isLoading ? "Сохранение..." : "Сохранить изменения"}
                            </button>
                        </div>
                    </div>

                    {(formError || error) && (
                        <div className="alert alert--danger">
                            {formError || error}
                        </div>
                    )}

                    {students.length === 0 && (
                        <div className="empty-state">
                            <h2 className="empty-state__title">
                                Ученики не найдены
                            </h2>

                            <p className="empty-state__text">
                                В этом классе пока нет учеников для добавления в мероприятие.
                            </p>
                        </div>
                    )}

                    {students.length > 0 && (
                        <div className="students-select-list">
                            {students.map((student) => {
                                const isSelected = selectedUserIds.includes(student.Id);
                                const wasInitiallySelected = initialSelectedUserIds.includes(student.Id);

                                return (
                                    <div
                                        key={student.Id}
                                        className={
                                            isSelected
                                                ? "student-select-card student-select-card--active"
                                                : "student-select-card"
                                        }
                                        role="button"
                                        tabIndex={0}
                                        onClick={() => toggleUser(student.Id)}
                                        onKeyDown={(event) => {
                                            if (event.key === "Enter" || event.key === " ") {
                                                event.preventDefault();
                                                toggleUser(student.Id);
                                            }
                                        }}
                                    >
                                        <div className="student-select-card__main">
                                            <div className="student-select-card__checkbox">
                                                <input
                                                    type="checkbox"
                                                    checked={isSelected}
                                                    onChange={() => toggleUser(student.Id)}
                                                    onClick={(event) => event.stopPropagation()}
                                                    disabled={isLoading}
                                                />
                                            </div>

                                            <div className="student-select-card__avatar">
                                                {student.Name[0]}
                                                {student.LastName[0]}
                                            </div>

                                            <div className="student-select-card__info">
                                                <h3 className="student-select-card__name">
                                                    {student.Name} {student.LastName}
                                                </h3>

                                                <p className="student-select-card__meta">
                                                    Логин: {student.Login}
                                                </p>

                                                {wasInitiallySelected && (
                                                    <p className="student-select-card__meta">
                                                        Уже участвует в мероприятии
                                                    </p>
                                                )}
                                            </div>
                                        </div>

                                        <div className="student-select-card__rating">
                                            <span className="student-select-card__rating-label">
                                                Рейтинг
                                            </span>

                                            <span className="student-select-card__rating-value">
                                                {student.Rating}
                                            </span>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>
            </section>
        </main>
    );
}