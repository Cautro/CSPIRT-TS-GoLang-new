import { type FormEvent, useEffect, useState } from "react";
import type { ClassType } from "../../../../shared/entities/class/types/class_types.ts";
import {
    addEventSchema,
    type addEventType,
} from "../../../../shared/entities/events/api/events_api.ts";

interface Props {
    isOpen: boolean;
    onClose: () => void;
    onEventAdd: (dto: addEventType) => Promise<void>;
    classes: ClassType[];
}

export function AddEventModal({
                                  isOpen,
                                  onClose,
                                  onEventAdd,
                                  classes,
                              }: Props) {
    const [formError, setFormError] = useState<string | null>(null);
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

    const selectedClassNames = classes
        .filter((item) => selectedClassIds.includes(item.Id))
        .map((item) => item.Name);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        setFormError(null);

        const formData = new FormData(event.currentTarget);

        const title = String(formData.get("title") ?? "").trim();
        const description = String(formData.get("description") ?? "").trim();
        const startedAtRaw = String(formData.get("startedAt") ?? "").trim();
        const rating = Number(formData.get("rating") ?? "");

        const dto = {
            Title: title,
            Description: description,
            StartedAt: startedAtRaw.replace("T", " "),
            Classes: selectedClassIds,
            RatingReward: rating,
        };

        const parsed = addEventSchema.safeParse(dto);

        if (!parsed.success) {
            console.log(parsed.error.issues);
            setFormError("Проверьте правильность заполнения полей");
            return;
        }

        try {
            setIsSubmitting(true);
            await onEventAdd(parsed.data);

            setSelectedClassIds([]);
            setIsClassesDropdownOpen(false);
            onClose();
        } catch (error) {
            setFormError(
                error instanceof Error
                    ? error.message
                    : "Ошибка при добавлении мероприятия"
            );
        } finally {
            setIsSubmitting(false);
        }
    }

    return (
        <div className="modal-backdrop" onMouseDown={onClose}>
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
                        onClick={onClose}
                        aria-label="Закрыть модальное окно"
                    >
                        ×
                    </button>
                </div>

                <form className="form" onSubmit={handleSubmit}>
                    <div className="modal__body">
                        {formError && (
                            <div className="alert alert--danger">
                                {formError}
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
                                    {selectedClassNames.length > 0
                                        ? selectedClassNames.join(", ")
                                        : "Выберите классы"}
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
                            onClick={onClose}
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