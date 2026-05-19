import {useEffect, useMemo, useState} from "react";
import {useNavigate, useParams} from "react-router-dom";

import type {UserType} from "../../../../shared/entities/user/types/user_types.ts";
import {addEventPlayersSchema} from "../../../../shared/entities/events/api/events_api.ts";
import {useEventStore} from "../../store/event_store.ts";
import {useClassStore} from "../../../class/store/class_store.ts";
import type {ClassType} from "../../../../shared/entities/class/types/class_types.ts";

export function EventClassPlayersPage() {
    const navigate = useNavigate();

    const {eventId, classId} = useParams<{ eventId: string, classId: string }>();

    const numericEventId = eventId ? Number(eventId) : null;
    const numericClassId = classId ? Number(classId) : null;

    const event = useEventStore((state) => state.event);
    const status = useEventStore((state) => state.status);
    const error = useEventStore((state) => state.error);
    const [classItem, setClassItem] = useState<ClassType | null>(null)

    const getEventById = useEventStore((state) => state.getEventById);
    const addPlayersToEvent = useEventStore((state) => state.addPlayersToEvent);
    const removePlayersFromEvent = useEventStore((state) => state.removePlayersFromEvent);
    const getClassById = useClassStore((state) => state.getClassById)

    const [selectedUserIds, setSelectedUserIds] = useState<number[]>([]);
    const [formError, setFormError] = useState<string | null>(null);

    const isLoading = status === "loading";

    useEffect(() => {
        if (!numericEventId || Number.isNaN(numericEventId)) {
            return;
        }
        void getEventById(numericEventId);
    }, [numericEventId, getEventById]);

    useEffect(() => {
        if (!numericClassId || Number.isNaN(numericClassId)) {
            return;
        }

        let isMounted = true;

        async function loadClass() {
            setClassItem(null);

            const item = await getClassById(numericClassId);

            if (isMounted) {
                setClassItem(item);
            }
        }

        void loadClass();

        return () => {
            isMounted = false;
        };
    }, [getClassById, numericClassId]);


    const students = useMemo<UserType[]>(() => {
        if (!classItem) {
            return [];
        }

        return classItem.Members.filter((user) => {
            return user.Role === "User" || user.Role === "Helper";
        });
    }, [classItem]);

    const studentIdSet = useMemo(() => {
        return new Set(students.map((student) => student.Id));
    }, [students]);

    const eventPlayerIdSet = useMemo(() => {
        const ids = event?.Players;

        return new Set(ids);
    }, [event]);

    const initialSelectedUserIds = useMemo(() => {
        return students
            .filter((student) => eventPlayerIdSet.has(student.Id))
            .map((student) => student.Id);
    }, [students, eventPlayerIdSet]);

    useEffect(() => {
        setSelectedUserIds(initialSelectedUserIds);
    }, [initialSelectedUserIds]);

    if (!numericEventId || Number.isNaN(numericEventId)) {
        return (
            <main className="main">
                <section className="page">
                    <div className="empty-state">
                        <h2 className="empty-state__title">ID мероприятия не найден</h2>
                        <p className="empty-state__text">Некорректный адрес страницы.</p>

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

    if (!numericClassId || Number.isNaN(numericClassId)) {
        return (
            <main className="main">
                <section className="page">
                    <div className="empty-state">
                        <h2 className="empty-state__title">ID класса не найден</h2>
                        <p className="empty-state__text">Некорректный адрес страницы.</p>

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

    if (isLoading && (!event || !classItem)) {
        return (
            <main className="main">
                <section className="page">
                    <div className="class-list">
                        <div className="skeleton" style={{height: 120}}/>
                        <div className="skeleton" style={{height: 88}}/>
                        <div className="skeleton" style={{height: 88}}/>
                    </div>
                </section>
            </main>
        );
    }

    if (!event || !classItem) {
        return (
            <main className="main">
                <section className="page">
                    <div className="empty-state">
                        <h2 className="empty-state__title">Данные не найдены</h2>

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

        if (!numericEventId) {
            return;
        }

        setFormError(null);

        const initialSet = new Set(initialSelectedUserIds);
        const selectedSet = new Set(selectedUserIds);

        const addedIds = selectedUserIds.filter((id) => {
            return studentIdSet.has(id) && !initialSet.has(id);
        });

        const removedIds = initialSelectedUserIds.filter((id) => {
            return studentIdSet.has(id) && !selectedSet.has(id);
        });

        if (addedIds.length === 0 && removedIds.length === 0) {
            setFormError("Изменений нет");
            return;
        }

        if (addedIds.length > 0 && numericEventId) {
            const parsedAdd = addEventPlayersSchema.safeParse({
                playerIds: addedIds,
            });

            if (!parsedAdd.success) {
                console.log(parsedAdd.error.issues);
                setFormError("Некорректный список учеников для добавления");
                return;
            }

            await addPlayersToEvent(numericEventId, parsedAdd.data);
        }

        if (removedIds.length > 0 && numericEventId) {
            const parsedRemove = addEventPlayersSchema.safeParse({
                playerIds: removedIds,
            });

            if (!parsedRemove.success) {
                console.log(parsedRemove.error.issues);
                setFormError("Некорректный список учеников для удаления");
                return;
            }

            await removePlayersFromEvent(numericEventId, parsedRemove.data);
        }

        await getEventById(numericEventId);
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