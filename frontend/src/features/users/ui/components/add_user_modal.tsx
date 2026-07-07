import { type FormEvent, useEffect, useState } from "react";
import {type UserRole, UserRoles} from "../../../../shared/entities/user/types/user_types.ts";
import type {addUserValues} from "../../models/add_user_usecase.ts";
import {useClasses} from "../../../class/hooks/use_classes.ts";
import {useAddUser} from "../../hooks/use_add_user.ts";

interface AddUserModalProps {
    isOpen: boolean;
    onClose: () => void;
    onAddUser: () => void;
    classId?: number | null
}

export function AddUserModal({isOpen, onClose, onAddUser, classId = null}: AddUserModalProps) {
    const classes = useClasses().data;
    const {mutateAsync, error} = useAddUser()
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [selectedRole, setSelectedRole] = useState<UserRole>("User");
    const normalizedSelectedRole = selectedRole.toLowerCase();
    const shouldShowClass = normalizedSelectedRole === "user" || normalizedSelectedRole === "helper";
    
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
        
        const form: addUserValues = {
            name: String(formData.get("name") ?? "").trim(),
            lastname: String(formData.get("lastName") ?? "").trim(),
            password: String(formData.get("password") ?? "").trim(),
            classId: shouldShowClass ? Number(String(formData.get("classId") ?? "")) : 0,
            login: String(formData.get("login") ?? "").trim(),
            role: String(formData.get("role") ?? "User").trim() as UserRole,
        };
        
        try {
            setIsSubmitting(true);
            await mutateAsync(form)
            onAddUser();
        } finally {
            setIsSubmitting(false);
        }
    }

    if (!classes) {
        return (
            <div className="modal-backdrop" onMouseDown={() => {onClose();;}}>
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
                                Не удалось найти классы, для добавления нового пользователя в один из них
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
                            Добавление пользователя
                        </h2>

                        <p className="modal__description">
                            Заполните данные нового пользователя системы.
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
                        {error && (
                            <div className="alert alert--danger">
                                {error.message}
                            </div>
                        )}

                        <div className="form-row">
                            <div className="field">
                                <label className="field__label" htmlFor="userName">
                                    Имя
                                </label>
                                <input
                                    id="userName"
                                    name="name"
                                    className="input"
                                    type="text"
                                    placeholder="Например: Иван"
                                    minLength={2}
                                    maxLength={20}
                                    required
                                />
                            </div>

                            <div className="field">
                                <label className="field__label" htmlFor="userLastName">
                                    Фамилия
                                </label>
                                <input
                                    id="userLastName"
                                    name="lastName"
                                    className="input"
                                    type="text"
                                    placeholder="Например: Петров"
                                    minLength={2}
                                    maxLength={20}
                                    required
                                />
                            </div>
                        </div>

                        <div className="form-row">
                            <div className="field">
                                <label className="field__label" htmlFor="userLogin">
                                    Логин
                                </label>
                                <input
                                    id="userLogin"
                                    name="login"
                                    className="input"
                                    type="text"
                                    placeholder="Например: ivan.petrov"
                                    autoComplete="username"
                                    minLength={2}
                                    maxLength={20}
                                    required
                                />
                            </div>

                            <div className="field">
                                <label className="field__label" htmlFor="userPassword">
                                    Пароль
                                </label>
                                <input
                                    id="userPassword"
                                    name="password"
                                    className="input"
                                    type="password"
                                    placeholder="Введите пароль"
                                    autoComplete="new-password"
                                    minLength={6}
                                    maxLength={35}
                                    required
                                />
                            </div>
                        </div>

                        <div className="form-row">
                            <div className="field">
                                <label className="field__label" htmlFor="userRole">
                                    Роль
                                </label>

                                <select
                                    id="userRole"
                                    name="role"
                                    className="select"
                                    value={selectedRole}
                                    onChange={(event) => setSelectedRole(event.target.value as UserRole)}
                                    required
                                >
                                    <option value="User">{UserRoles.User}</option>
                                    <option value="Helper">{UserRoles.Helper}</option>
                                    {!classId && <option value="Public">{UserRoles.Public}</option>}
                                    {!classId && (<option value="Admin">{UserRoles.Admin}</option>)}
                                    {!classId && (<option value="Owner">{UserRoles.Owner}</option>)}
                                </select>
                            </div>
                            
                            {shouldShowClass && (
                                <div className="field">
                                    <label className="field__label" htmlFor="userClass">
                                        Класс
                                    </label>

                                    {!classId ? (<select
                                        id="userClass"
                                        name="classId"
                                        className="select"
                                        defaultValue=""
                                        required={shouldShowClass}
                                    >
                                        <option value="" disabled>
                                            Выберите класс
                                        </option>

                                        {classes.map((item) => (
                                            <option key={item.Id} value={String(item.Id)}>
                                                {item.Name}
                                            </option>
                                        ))}
                                    </select>) : (<select
                                        id="userClass"
                                        name="classId"
                                        className="select"
                                        defaultValue={classId}
                                        disabled={true}
                                        required={shouldShowClass}
                                    >
                                        <option value="" disabled>
                                            Выберите класс
                                        </option>

                                        {classes.map((item) => (
                                            <option key={item.Id} value={String(item.Id)}>
                                                {item.Name}
                                            </option>
                                        ))}
                                    </select>)}
                                </div>
                            )}

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
                            {isSubmitting ? "Создание..." : "Создать пользователя"}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    );
}