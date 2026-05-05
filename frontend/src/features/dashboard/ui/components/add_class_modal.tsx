import { type FormEvent, useEffect, useState } from "react";
import {type UserType} from "../../../../shared/entities/user/types/user_types";
import {addClassDto, type addClassType} from "../../../../shared/entities/class/api/class_api.ts";

interface AddUserModalProps {
    isOpen: boolean;
    onClose: () => void;
    onAddClass: (dto: addClassType) => Promise<void>;
    staff: UserType[];
}

export function AddClassModal({isOpen, onClose, onAddClass, staff}: AddUserModalProps) {
    const [formError, setFormError] = useState<string | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);
    
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

        const name = String(formData.get("name") ?? "").trim();
        const teacher = String(formData.get("teacher") ?? "").trim();

        const dto = {
            Name: name,
            TeacherLogin: teacher,
        };

        const parsed = addClassDto.safeParse(dto);

        if (!parsed.success) {
            console.log(parsed.error.format());
            setFormError("Проверьте правильность заполнения полей");
            return;
        }

        try {
            setIsSubmitting(true);
            await onAddClass(parsed.data);
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
                            Добавление нового класса
                        </h2>

                        <p className="modal__description">
                            Выберите название и классного руководителя нового класса
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
                                <label className="field__label" htmlFor="className">
                                    Имя
                                </label>
                                <input
                                    id="userName"
                                    name="name"
                                    className="input"
                                    type="text"
                                    placeholder="Например: 10А"
                                    maxLength={20}
                                    required
                                />
                            </div>

                            <div className="field">
                                <label className="field__label" htmlFor="teacher">
                                    Классный руководитель
                                </label>

                                <select
                                    id="userRole"
                                    name="teacher"
                                    className="select"
                                    defaultValue=""
                                >

                                    <option value="" disabled>
                                        Выберите учителя
                                    </option>

                                    {staff.map((item) => (
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
                            {isSubmitting ? "Добавление..." : "Добавить класс"}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    );
}