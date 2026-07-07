import { type FormEvent, useEffect, useState } from "react";
import { createPortal } from "react-dom";
import {type ScheduleAddFormValues, ScheduleAddLessonUsecase} from "../../models/add_schedule_usecase.ts";
import {dayLabels, type ScheduleDay} from "../../../../shared/entities/schedule/types/schedule_types.ts";
import {useStaff} from "../../../users/hooks/use_staff.ts";
import {UseAddSchedule} from "../../hooks/use_add_schedule.ts";


interface AddScheduleLessonModalProps {
    isOpen: boolean;
    onClose: () => void;
    classId: number;
    dayOfWeek: ScheduleDay;
    onAdded: () => Promise<void> | void;
    type: "base" | "current" | "planned"
}

export function AddScheduleLessonModal({isOpen, onClose, classId, dayOfWeek, onAdded, type}: AddScheduleLessonModalProps) {
    const addSchedule = UseAddSchedule()
    const getTeachers = useStaff();
    const teachers = getTeachers.data;
    const error = addSchedule.error?.message;

    const [isSubmitting, setIsSubmitting] = useState(false);

    const isLoading = isSubmitting || addSchedule.isPending;

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

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        const formData = new FormData(event.currentTarget);

        const form: ScheduleAddFormValues = {
            subject: String(formData.get("subject") ?? "").trim(),
            teacher_id: Number(formData.get("teacher_id")),
            start_time: String(formData.get("start_time") ?? "").trim(),
            end_time: String(formData.get("end_time") ?? "").trim(),
            room: Number(formData.get("room")),
            lesson_number: Number(formData.get("lesson_number")),
            day_of_week: dayOfWeek,
            class_id: classId,
        };

        try {
            setIsSubmitting(true);

            const dto = ScheduleAddLessonUsecase(form);
            addSchedule.mutateAsync({form: dto, type: type});
            onAdded();
        } finally {
            setIsSubmitting(false);
        }
    }

    return createPortal(
        <div className="modal-backdrop" onMouseDown={onClose}>
            <section
                className="modal modal--wide"
                role="dialog"
                aria-modal="true"
                aria-labelledby="add-schedule-lesson-title"
                onMouseDown={(event) => event.stopPropagation()}
            >
                <div className="modal__header">
                    <div>
                        <h2 className="modal__title" id="add-schedule-lesson-title">
                            Добавление урока
                        </h2>

                        <p className="modal__description">
                            Добавление нового урока в расписание.
                        </p>
                    </div>

                    <button
                        className="modal__close"
                        type="button"
                        onClick={onClose}
                        disabled={isLoading}
                        aria-label="Закрыть модальное окно"
                    >
                        ×
                    </button>
                </div>

                <form className="form" onSubmit={handleSubmit}>
                    <div className="modal__body">
                        {error && <div className="alert alert--danger">{error}</div>}

                        <div className="schedule-edit-summary">
                            <div className="schedule-edit-summary__number">+</div>

                            <div>
                                <div className="schedule-edit-summary__title">
                                    Новый урок
                                </div>

                                <div className="schedule-edit-summary__text">
                                    Урок будет добавлен на {dayLabels[dayOfWeek]}
                                </div>
                            </div>
                        </div>

                        <div className="form-row">
                            <div className="field">
                                <label className="field__label" htmlFor="lessonSubject">
                                    Предмет
                                </label>

                                <input
                                    id="lessonSubject"
                                    name="subject"
                                    className="input"
                                    type="text"
                                    placeholder="Например: Математика"
                                    minLength={2}
                                    maxLength={20}
                                    required
                                />
                            </div>

                            <div className="field">
                                <label className="field__label" htmlFor="lessonNumber">
                                    Номер урока
                                </label>

                                <input
                                    id="lessonNumber"
                                    name="lesson_number"
                                    className="input"
                                    type="number"
                                    min={1}
                                    placeholder="Например: 1"
                                    required
                                />
                            </div>
                        </div>

                        <div className="form-row">
                            <div className="field">
                                <label className="field__label" htmlFor="lessonTeacher">
                                    Учитель
                                </label>

                                <select
                                    id="lessonTeacher"
                                    name="teacher_id"
                                    className="select"
                                    defaultValue=""
                                    required
                                >
                                    <option value="" disabled>
                                        Выберите учителя
                                    </option>

                                    {teachers?.map((teacher) => (
                                        <option key={teacher.Id} value={String(teacher.Id)}>
                                            {teacher.Name} {teacher.LastName}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <div className="field">
                                <label className="field__label" htmlFor="lessonRoom">
                                    Кабинет
                                </label>

                                <input
                                    id="lessonRoom"
                                    name="room"
                                    className="input"
                                    type="number"
                                    min={1}
                                    placeholder="Например: 101"
                                    required
                                />
                            </div>
                        </div>

                        <div className="form-row">
                            <div className="field">
                                <label className="field__label" htmlFor="lessonStartTime">
                                    Начало урока
                                </label>

                                <input
                                    id="lessonStartTime"
                                    name="start_time"
                                    className="input"
                                    type="time"
                                    step={60}
                                    required
                                />

                                <div className="field__hint">Формат: HH:mm</div>
                            </div>

                            <div className="field">
                                <label className="field__label" htmlFor="lessonEndTime">
                                    Конец урока
                                </label>

                                <input
                                    id="lessonEndTime"
                                    name="end_time"
                                    className="input"
                                    type="time"
                                    step={60}
                                    required
                                />

                                <div className="field__hint">Формат: HH:mm</div>
                            </div>
                        </div>
                    </div>

                    <div className="modal__footer">
                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={onClose}
                            disabled={isLoading}
                        >
                            Отмена
                        </button>

                        <button
                            className="btn btn--primary"
                            type="submit"
                            disabled={isLoading}
                        >
                            {isLoading ? "Добавление..." : "Добавить урок"}
                        </button>
                    </div>
                </form>
            </section>
        </div>,
        document.body,
    );
}