import {useEffect, useMemo, useRef, useState} from "react";
import {useNavigate, useParams} from "react-router-dom";
import type {UserType} from "../../../../shared/entities/user/types/user_types.ts";
import type {ClassType} from "../../../../shared/entities/class/types/class_types.ts";
import {addEventPlayersSchema} from "../../../../shared/entities/events/api/events_api.ts";
import {PageHeader} from "../../../../shared/ui/other/page_header.tsx";
import {UseEventById} from "../../hooks/use_event_by_id.ts";
import {useAddEventPlayers} from "../../hooks/use_add_event_players.ts";
import {useRemoveEventPlayers} from "../../hooks/use_remove_event_players.ts";
import {useClassId} from "../../../class/hooks/use_class_id.ts";

export function EventClassPlayersPage() {
    const navigate = useNavigate();

    const {eventId, classId} = useParams<{ eventId: string; classId: string; }>();
    
    const numericEventId = eventId ? Number(eventId) : null;
    const numericClassId = classId ? Number(classId) : null;

    const getEventById = UseEventById(numericEventId ?? 0);
    const getClassById = useClassId(numericClassId ?? 0);

    const addPlayersToEvent = useAddEventPlayers();
    const removePlayersFromEvent = useRemoveEventPlayers();

    const event = getEventById.data;
    const classItem = getClassById.data as ClassType | null;

    const [selectedUserIds, setSelectedUserIds] = useState<number[]>([]);
    const [formError, setFormError] = useState<string | null>(null);

    const initializedRef = useRef(false);

    const error =
        getEventById.error?.message ||
        addPlayersToEvent.error?.message ||
        removePlayersFromEvent.error?.message ||
        getClassById.error?.message ||
        null;

    const isLoading =
        getEventById.isLoading ||
        getClassById.isLoading ||
        addPlayersToEvent.isPending ||
        removePlayersFromEvent.isPending;

    const students = useMemo<UserType[]>(() => {
        if (!classItem) {
            return [];
        }

        return classItem.Members.filter((user) => {
            const normalizedRole = user.Role.toLowerCase();

            return normalizedRole === "user" || normalizedRole === "helper";
        });
    }, [classItem]);

    const studentIdSet = useMemo(() => {
        return new Set(students.map((student) => student.Id));
    }, [students]);

    const eventPlayerIdSet = useMemo(() => {
        return new Set(event?.Players ?? []);
    }, [event]);

    const initialSelectedUserIds = useMemo(() => {
        return students
            .filter((student) => eventPlayerIdSet.has(student.Id))
            .map((student) => student.Id);
    }, [students, eventPlayerIdSet]);

    useEffect(() => {
        if (!initializedRef.current && !getEventById.isLoading && !getClassById.isLoading) {
            setSelectedUserIds(initialSelectedUserIds);
            initializedRef.current = true;
        }
    }, [initialSelectedUserIds, getEventById.isLoading, getClassById.isLoading]);

    if (!numericEventId || Number.isNaN(numericEventId)) {
        return (
            <main className="main">
                <section className="page">
                    <div className="empty-state">
                        <h2 className="empty-state__title">
                            ID мероприятия не найден
                        </h2>

                        <p className="empty-state__text">
                            Некорректный адрес страницы.
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

    if (!numericClassId || Number.isNaN(numericClassId)) {
        return (
            <main className="main">
                <section className="page">
                    <div className="empty-state">
                        <h2 className="empty-state__title">
                            ID класса не найден
                        </h2>

                        <p className="empty-state__text">
                            Некорректный адрес страницы.
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

    if (isLoading && (!event || !classItem)) {
        return (
            <main className="main">
                <section className="page">
                    <div className="class-list">
                        <div
                            className="skeleton"
                            style={{height: 120}}
                        />

                        <div
                            className="skeleton"
                            style={{height: 88}}
                        />

                        <div
                            className="skeleton"
                            style={{height: 88}}
                        />
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

        if (addedIds.length > 0) {
            const parsedAdd = addEventPlayersSchema.safeParse({
                playerIds: addedIds,
            });

            if (!parsedAdd.success) {
                console.log(parsedAdd.error.issues);

                setFormError(
                    "Некорректный список учеников для добавления"
                );

                return;
            }

            await addPlayersToEvent.mutateAsync({
                id: numericEventId,
                dto: parsedAdd.data,
            });
        }

        if (removedIds.length > 0) {
            const parsedRemove = addEventPlayersSchema.safeParse({
                playerIds: removedIds,
            });

            if (!parsedRemove.success) {
                console.log(parsedRemove.error.issues);

                setFormError(
                    "Некорректный список учеников для удаления"
                );

                return;
            }

            await removePlayersFromEvent.mutateAsync({
                id: numericEventId,
                dto: parsedRemove.data,
            });
        }

        await getEventById.refetch();

        navigate(-1);
    }

    return (
        <main className="main">
            <section className="page">
                <div className="event-players-page">

                    <PageHeader
                        eyebrow={`Мероприятие #${event.ID}`}
                        title={"Участники мероприятия"}
                        description={`${event.Title} · ${classItem.Name} класс`}
                        hasBackButton={true}
                    />

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
                                {isLoading
                                    ? "Сохранение..."
                                    : "Сохранить изменения"}
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
                                В этом классе пока нет учеников
                                для добавления в мероприятие.
                            </p>
                        </div>
                    )}

                    {students.length > 0 && (
                        <div className="students-select-list">
                            {students.map((student) => {
                                const isSelected =
                                    selectedUserIds.includes(student.Id);

                                const wasInitiallySelected =
                                    initialSelectedUserIds.includes(student.Id);

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
                                            if (
                                                event.key === "Enter" ||
                                                event.key === " "
                                            ) {
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
                                                    onClick={(event) => {
                                                        event.stopPropagation();
                                                    }}
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
                                                        Уже участвует
                                                        в мероприятии
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