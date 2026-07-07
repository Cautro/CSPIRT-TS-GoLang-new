import { type FormEvent, useEffect, useState } from "react";
import {useStaff} from "../../../../features/users/hooks/use_staff.ts";
interface ChangeTeacherProps {
    isOpen: boolean;
    onClose: () => void;
    onChangeTeacher: (teacher: string) => Promise<void>;
    className: string;
}

export function ChangeTeacherModal({isOpen, onClose, onChangeTeacher, className}: ChangeTeacherProps) {
    const [formError, setFormError] = useState<string | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const staff = useStaff().data

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
        setFormError(null);

        const formData = new FormData(event.currentTarget);

        const teacher = String(formData.get("teacher") ?? "");
        
        try {
            setIsSubmitting(true);
            await onChangeTeacher(teacher);
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
                aria-labelledby="add-user-modal-title"
                onMouseDown={(event) => event.stopPropagation()}
            >
                <div className="modal__header">
                    <div>
                        <h2 className="modal__title" id="add-user-modal-title">
                            Изменение классного руководителя
                        </h2>

                        <p className="modal__description">
                            Выберите нового классного руководителя для {className} класса
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
                                <label className="field__label" htmlFor="teacher">
                                    Классный руководитель
                                </label>

                                <select
                                    id="userRole"
                                    name="teacher"
                                    className="select"
                                    defaultValue=""
                                    required
                                >
                                    
                                    <option value="" disabled>
                                        Выберите учителя
                                    </option>
                                    
                                    {staff?.map((item) => (
                                        <option value={item.Login}>{item.Name} {item.LastName}</option>
                                    ))}
                                </select>
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
                            {isSubmitting ? "Изменение..." : "Изменить"}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    );
}