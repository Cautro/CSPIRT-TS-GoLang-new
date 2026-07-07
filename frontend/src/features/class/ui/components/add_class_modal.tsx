import { type FormEvent, useEffect, useState } from "react";
import type {addClassFormValues} from "../../models/add_class_usecase.ts";
import {useAddClass} from "../../hooks/use_add_class.ts";
import {useStaff} from "../../../users/hooks/use_staff.ts";

interface AddUserModalProps {
    isOpen: boolean;
    onClose: () => void;
    onAddClass: () => Promise<void>;
}

export function AddClassModal({isOpen, onClose, onAddClass}: AddUserModalProps) {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const getStaff = useStaff();
    const staff = getStaff.data
    const {mutateAsync, error, isError} = useAddClass()
    
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

        const dto: addClassFormValues = {
            name: String(formData.get("name") ?? "").trim(),
            teacher_login: String(formData.get("teacher") ?? "").trim(),
        };

        try {
            setIsSubmitting(true);
            await mutateAsync(dto);
            onAddClass();
        } finally {
            setIsSubmitting(false);
        }
    }

    if (!staff) {
        return (
            <div className="modal-backdrop" onMouseDown={() => {onClose();}}>
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
                                Учителя не найдены
                            </h2>

                            <p className="modal__description">
                                Не удалось найти учителей для назначения классным руководителем
                            </p>
                        </div>
                    </div>
                </section>
            </div>
        )
    }

    return (
        <div className="modal-backdrop" onMouseDown={() => {onClose();}}>
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
                        onClick={() => {onClose();}}
                        aria-label="Закрыть модальное окно"
                    >
                        ×
                    </button>
                </div>

                <form className="form" onSubmit={handleSubmit}>
                    <div className="modal__body">
                        {isError && (
                            <div className="alert alert--danger">
                                {error.message}
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
                                        <option key={item.Id} value={item.Login}>{item.Name} {item.LastName}</option>
                                    ))}
                                </select>
                            </div>

                        </div>

                    </div>

                    <div className="modal__footer">
                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => {onClose();}}
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