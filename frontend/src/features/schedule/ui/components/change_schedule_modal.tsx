import { type FormEvent, useEffect, useState } from "react";
import type { ScheduleChangeFormValues } from "../../models/schedule_change_usecase.ts";
import type {ScheduleLessonType} from "../../../../shared/entities/schedule/types/schedule_types.ts";
import { createPortal } from "react-dom";
import {ConfirmModal} from "../../../../shared/ui/modals/confirm_modal.tsx";
import {useScheduleStore} from "../../store/schedule_store.ts";

interface ChangeScheduleLessonModalProps {
    isOpen: boolean;
    onClose: () => void;
    lesson: ScheduleLessonType | null;
    onChanged: () => Promise<void> | void;
    type: "base" | "current" | "planned"
}

export function ChangeScheduleLessonModal({isOpen, onClose, lesson, onChanged, type,}: ChangeScheduleLessonModalProps) {
    const error = useScheduleStore((state) => state.error);
    const status = useScheduleStore((state) => state.status);
    const teachers = useScheduleStore((state) => state.teachers);
    const changeSchedule = useScheduleStore((state) => state.changeSchedule);
    const getTeachers = useScheduleStore((state) => state.getTeachers);
    const [isDeleteLessonModalOpen, setIsDeleteLessonModalOpen] = useState(false);
    const deleteSchedule = useScheduleStore((state) => state.deleteSchedule)
    
    const [isSubmitting, setIsSubmitting] = useState(false);
    const isLoading = isSubmitting || status === "loading";

    useEffect(() => {
        void getTeachers();
    }, [getTeachers]);
    
    useEffect(() => {
        if (!isOpen) {
            return;
        }

        function handleEscape(event: KeyboardEvent) {
            if (event.key === "Escape") {
                useScheduleStore.setState({error: null});
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

    if (!isOpen || !lesson) {
        return null;
    }

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        if (!lesson) {
            return;
        }

        const formData = new FormData(event.currentTarget);

        const form: ScheduleChangeFormValues = {
            subject: String(formData.get("subject") ?? "").trim(),
            teacher_id: Number(formData.get("teacher_id")),
            start_time: String(formData.get("start_time") ?? "").trim(),
            end_time: String(formData.get("end_time") ?? "").trim(),
            room: Number(formData.get("room") ?? ""),
        };

        try {
            setIsSubmitting(true);

            const success = await changeSchedule(lesson.Id, form, type);

            if (success) {
                await onChanged();
                useScheduleStore.setState({error: null});
                onClose();
            }
        } finally {
            setIsSubmitting(false);
        }
    }

    return createPortal(
        <div className="modal-backdrop" onMouseDown={() => {onClose(); useScheduleStore.setState({error: null});}}>
            <section
                className="modal modal--wide"
                role="dialog"
                aria-modal="true"
                aria-labelledby="change-schedule-lesson-title"
                onMouseDown={(event) => event.stopPropagation()}
            >
                <div className="modal__header">
                    <div>
                        <h2 className="modal__title" id="change-schedule-lesson-title">
                            Изменение урока
                        </h2>

                        <p className="modal__description">
                            Измените данные урока №{lesson.LessonNumber} за выбранный день.
                        </p>
                    </div>

                    <button
                        className="modal__close"
                        type="button"
                        onClick={()=> {onClose(); useScheduleStore.setState({error: null});}}
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
                            <div className="schedule-edit-summary__number">
                                {lesson.LessonNumber}
                            </div>

                            <div>
                                <div className="schedule-edit-summary__title">
                                    {lesson.Subject}
                                </div>

                                <div className="schedule-edit-summary__text">
                                    {lesson.StartTime} — {lesson.EndTime}
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
                                    defaultValue={lesson.Subject}
                                    placeholder="Например: Математика"
                                    minLength={2}
                                    maxLength={50}
                                    required
                                />
                            </div>

                            <div className="field">
                                <label className="field__label" htmlFor="lessonTeacher">
                                    Учитель
                                </label>

                                <select
                                    id="lessonTeacher"
                                    name="teacher_id"
                                    className="select"
                                    defaultValue={String(lesson.TeacherID)}
                                    required
                                >
                                    <option value="" disabled>
                                        Выберите учителя
                                    </option>

                                    {teachers !== null && (teachers.map((teacher) => (
                                        <option key={teacher.Id} value={String(teacher.Id)}>
                                            {teacher.Name} {teacher.LastName}
                                        </option>
                                    )))}
                                </select>
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
                                    defaultValue={lesson.StartTime}
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
                                    defaultValue={lesson.EndTime}
                                    step={60}
                                    required
                                />

                                <div className="field__hint">Формат: HH:mm</div>
                            </div>
                        </div>

                        <div className="form-row">
                            <div className="field">
                                <label className="field__label" htmlFor="lessonRoom">
                                    Кабинет
                                </label>

                                <input
                                    id="lessonRoom"
                                    name="room"
                                    className="input"
                                    type="number"
                                    defaultValue={String(lesson.Room)}
                                    placeholder="Например: 101"
                                    minLength={1}
                                    maxLength={20}
                                    required
                                />
                            </div>
                        </div>
                    </div>

                    <div className="modal__footer">
                        <button
                            className="btn btn--danger"
                            type="button"
                            onClick={() => setIsDeleteLessonModalOpen(true)}
                        >
                            Отменить урок
                        </button>
                        
                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => {onClose(); useScheduleStore.setState({error: null});}}
                            disabled={isLoading}
                        >
                            Отмена
                        </button>

                        <button
                            className="btn btn--primary"
                            type="submit"
                            disabled={isLoading}
                        >
                            {isLoading ? "Сохранение..." : "Сохранить изменения"}
                        </button>
                    </div>
                </form>

                <ConfirmModal
                    isOpen={isDeleteLessonModalOpen}
                    onClose={() => setIsDeleteLessonModalOpen(false)}
                    onConfirm={async () => {
                        await deleteSchedule(lesson.Id, type);
                        onChanged();
                    }}
                    isDanger={true}
                    title={"Отменить урок"}
                    content={`Отменить проведение урока в этот день?`}
                    buttonContent={"Отменить"}
                />
                
            </section>
        </div>,
        document.body
    );
}