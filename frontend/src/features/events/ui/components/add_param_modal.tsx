import { type FormEvent, useEffect, useState } from "react";
import type {AddParamFormValues} from "../../models/add_param_usecase.ts";
import {UseAddParam} from "../../hooks/use_add_param.ts";
import type {ClassType} from "../../../../shared/entities/class/types/class_types.ts";

interface AddUserModalProps {
    isOpen: boolean;
    onClose: () => void;
    Classes: ClassType[]
    EventID: number;
    OnAdd: () => void;
}

export function AddParamModal({isOpen, onClose, Classes, EventID, OnAdd}: AddUserModalProps) {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const {mutateAsync, error, isError} = UseAddParam()
    const [isBaranka, setIsBaranka] = useState(true);

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

        const dto: AddParamFormValues = {
            reason: String(formData.get("reason") ?? "").trim(),
            rating: Number(formData.get("reward")),
            class_id: Number(formData.get("class") ?? ""),
            event_id: EventID,
        };

        try {
            setIsSubmitting(true);
            await mutateAsync(dto);
            OnAdd();
        } finally {
            setIsSubmitting(false);
        }
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
                            Изменение награды для класса
                        </h2>

                        <p className="modal__description">
                            Введите описание для награды и выберите значение присуждаемого рейтинга
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
                                    Описание
                                </label>
                                <input
                                    id="reason"
                                    name="reason"
                                    className="input"
                                    type="text"
                                    placeholder="Например: Победитель"
                                    maxLength={200}
                                    required
                                />
                            </div>

                            <div className="field" style={{marginTop: "-25px"}}>
                                <div className="field__top">
                                    <label className="field__label" htmlFor="reward">
                                        Награда
                                    </label>

                                    <div className="reward-mode">
                                        <button
                                            type="button"
                                            className={`reward-mode__button ${isBaranka ? "active" : ""}`}
                                            onClick={() => setIsBaranka(true)}
                                        >
                                            Готовые
                                        </button>

                                        <button
                                            type="button"
                                            className={`reward-mode__button ${!isBaranka ? "active" : ""}`}
                                            onClick={() => setIsBaranka(false)}
                                        >
                                            Вручную
                                        </button>
                                    </div>
                                </div>

                                {isBaranka ? (
                                    <select
                                        id="reward"
                                        name="reward"
                                        className="select"
                                        required
                                    >
                                        <option value="50">50 Рейтинга</option> 🥨
                                        <option value="100">100 Рейтинга</option>
                                        <option value="250">250 Рейтинга</option> ⭐
                                        <option value="500">500 Рейтинга</option>  🚀
                                    </select>
                                ) : (
                                    <input
                                        id="reward"
                                        name="reward"
                                        className="input"
                                        type="number"
                                        max={5000}
                                        placeholder="Введите награду"
                                        required
                                    />
                                )}
                            </div>
                            
                            <div className="field">
                                <label className="field__label" htmlFor="teacher">
                                    Класс
                                </label>

                                <select
                                    id="class"
                                    name="class"
                                    className="select"
                                    defaultValue=""
                                >

                                    <option value="" disabled>
                                        Выберите класс
                                    </option>

                                    {Classes.map((item) => (
                                        <option key={item.Id} value={item.Id}>{item.Name} Класс</option>
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
                            {isSubmitting ? "Добавление..." : "Добавить награду"}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    );
}