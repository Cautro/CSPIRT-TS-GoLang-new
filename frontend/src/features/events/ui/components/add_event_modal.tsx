import { type FormEvent, useEffect, useState } from "react";
import {type addEventFormValues, addEventUsecase} from "../../models/add_event_usecase.ts";
import {useAddEvent} from "../../hooks/use_add_event.ts";
import {useClasses} from "../../../class/hooks/use_classes.ts";
import {UseEvents} from "../../hooks/use_events.ts";

interface Props {
    isOpen: boolean;
    onClose: () => void;
    onEventAdd: () => Promise<void>;
}

export function AddEventModal({isOpen, onClose, onEventAdd,}: Props) {
    const [error, setError] = useState<string | null>(null);
    const classes = useClasses().data
    const addEvent = useAddEvent()
    const getEvents = UseEvents();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [selectedClassIds, setSelectedClassIds] = useState<number[]>([]);
    const [isClassesDropdownOpen, setIsClassesDropdownOpen] = useState(false);
    
    useEffect(() => {
        if (!isOpen) {
            return;
        }

        function handleEscape(event: KeyboardEvent) {
            if (event.key === "Escape") {
                onClose();
            }
        }

        document.addEventListener("keydown", handleEscape);
        document.body.style.overflow = "hidden";

        return () => {
            document.removeEventListener("keydown", handleEscape);
            document.body.style.overflow = "";
        };
    }, [isOpen, onClose]);

    if (!isOpen) {
        return null;
    }

    function toggleClassId(classId: number) {
        setSelectedClassIds((prev) => {
            if (prev.includes(classId)) {
                return prev.filter((id) => id !== classId);
            }

            return [...prev, classId];
        });
    }

    const selectedClassNames = classes ? classes.filter((item) => selectedClassIds.includes(item.Id)).map((item) => item.Name) : [];
    
    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        
        if (selectedClassIds.length === 0) {
            setError("Выберите классы, участвующие в мероприятии");
            return;
        }
        

        const formData = new FormData(event.currentTarget);
        const startedAtRaw = String(formData.get("startedAt") ?? "").trim();
        
        const form: addEventFormValues = {
            title: String(formData.get("title") ?? "").trim(),
            description: String(formData.get("description") ?? "").trim(),
            started_at: startedAtRaw.replace("T", " "),
            classes: selectedClassIds,
            rating_reward: Number(formData.get("rating") ?? ""),
        };

        try {
            setIsSubmitting(true);
            const dto = await addEventUsecase(form);
            await addEvent.mutateAsync(dto);
            await onEventAdd();
            setSelectedClassIds([]);
            setIsClassesDropdownOpen(false);
            await getEvents.refetch();
            onClose();
        } catch (e) {
            setError(e instanceof Error ? e.message : "Неизвестная ошибка")
        }
        finally {
            setIsSubmitting(false);
        }
    }
    
    if (!classes) {
        return (
            <div className="modal-backdrop" onMouseDown={() => {onClose(); setError(null);}}>
                <section
                    className="modal modal--wide"
                    role="dialog"
                    aria-modal="true"
                    aria-labelledby="add-event-modal-title"
                    onMouseDown={(event) => event.stopPropagation()}
                >
                    <div className="modal__header">
                        <div>
                            <h2 className="modal__title" id="add-event-modal-title">
                                Классы не найдены
                            </h2>

                            <p className="modal__description">
                                Не удалось найти классы, для выбора их как участвующих в мероприятии
                            </p>
                        </div>
                    </div>
                </section>
            </div>
        )
    }

    return (
        <div className="modal-backdrop" onMouseDown={() => {onClose(); setError(null);}}>
            <section
                className="modal modal--wide"
                role="dialog"
                aria-modal="true"
                aria-labelledby="add-event-modal-title"
                onMouseDown={(event) => event.stopPropagation()}
            >
                <div className="modal__header">
                    <div>
                        <h2 className="modal__title" id="add-event-modal-title">
                            Добавление нового мероприятия
                        </h2>

                        <p className="modal__description">
                            Укажите данные о новом мероприятии в школе.
                        </p>
                    </div>

                    <button
                        className="modal__close"
                        type="button"
                        onClick={() => {onClose(); setError(null);}}
                        aria-label="Закрыть модальное окно"
                    >
                        ×
                    </button>
                </div>

                <form className="form" onSubmit={handleSubmit}>
                    <div className="modal__body">
                        {error && (
                            <div className="alert alert--danger">
                                {error}
                            </div>
                        )}

                        <div className="form-row">
                            <div className="field">
                                <label htmlFor="eventTitle" className="field__label">
                                    Название мероприятия
                                </label>

                                <input
                                    id="eventTitle"
                                    name="title"
                                    type="text"
                                    className="input"
                                    minLength={4}
                                    maxLength={64}
                                    placeholder="Например: Школьная олимпиада"
                                    required
                                />
                            </div>

                            <div className="field">
                                <label htmlFor="eventStartedAt" className="field__label">
                                    Время начала мероприятия
                                </label>

                                <input
                                    id="eventStartedAt"
                                    name="startedAt"
                                    type="datetime-local"
                                    className="input"
                                    required
                                />
                            </div>
                        </div>

                        <div className="field">
                            <label htmlFor="eventDescription" className="field__label">
                                Описание мероприятия
                            </label>

                            <textarea
                                id="eventDescription"
                                name="description"
                                className="input"
                                minLength={10}
                                maxLength={1000}
                                placeholder="Кратко опишите мероприятие"
                                rows={4}
                                required
                            />
                        </div>

                        <div className="field">
                            <label htmlFor="eventRating" className="field__label">
                                Вознаграждение за участие
                            </label>

                            <input
                                id="eventRating"
                                name="rating"
                                type="number"
                                className="input"
                                placeholder="Например: 50"
                                min={1}
                                max={5000}
                                step={1}
                                required
                            />
                        </div>

                        <div className="field">
                            <label className="field__label">
                                Классы, участвующие в мероприятии
                            </label>

                            <div className="multi-select">
                                <button
                                    className="select multi-select__button"
                                    type="button"
                                    onClick={() =>
                                        setIsClassesDropdownOpen((value) => !value)
                                    }
                                >
                                    {selectedClassNames.length > 0 ? selectedClassNames.join(", ") : "Выберите классы"}
                                </button>

                                {isClassesDropdownOpen && (
                                    <div className="multi-select__dropdown">
                                        {classes.length === 0 && (
                                            <div className="multi-select__empty">
                                                Классы не найдены
                                            </div>
                                        )}

                                        {classes.map((item) => (
                                            <label
                                                key={item.Id}
                                                className="multi-select__option"
                                            >
                                                <input
                                                    type="checkbox"
                                                    checked={selectedClassIds.includes(item.Id)}
                                                    onChange={() => toggleClassId(item.Id)}
                                                />

                                                <span>{item.Name}</span>
                                            </label>
                                        ))}
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    <div className="modal__footer">
                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => {onClose(); setError(null);}}
                            disabled={isSubmitting}
                        >
                            Отмена
                        </button>

                        <button
                            className="btn btn--primary"
                            type="submit"
                            disabled={isSubmitting}
                        >
                            {isSubmitting ? "Добавление..." : "Добавить мероприятие"}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    );
}