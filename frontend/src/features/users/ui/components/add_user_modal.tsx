import { type FormEvent, useEffect, useState } from "react";
import {type UserRole, UserRoles} from "../../../../shared/entities/user/types/user_types.ts";
import type {addUserValues} from "../../models/add_user_usecase.ts";
import {useUsersStore} from "../../store/users_store.ts";
import {useEventStore} from "../../../events/store/event_store.ts";
import {useClassStore} from "../../../class/store/class_store.ts";

interface AddUserModalProps {
    isOpen: boolean;
    onClose: () => void;
    onAddUser: () => void;
}

export function AddUserModal({isOpen, onClose, onAddUser}: AddUserModalProps) {
    const error = useUsersStore((state) => state.error);
    const classes = useClassStore((state) => state.classes);
    const getClasses = useClassStore((state) => state.getClasses);
    const addUser = useUsersStore((state) => state.addUser);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [selectedRole, setSelectedRole] = useState<UserRole>("User");
    const shouldShowClass = selectedRole === "User" || selectedRole === "Helper";

    useEffect(() => {
        void getClasses();
    }, [getClasses]);
    
    useEffect(() => {
        if (!isOpen) {
            return;
        }

        function handleEscape(event: KeyboardEvent) {
            if (event.key === "Escape") {
                useUsersStore.setState({error: null});
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
            const response = await addUser(form);
            if (response) {
                onAddUser();
            }
        } finally {
            setIsSubmitting(false);
        }
    }

    if (!classes) {
        return (
            <div className="modal-backdrop" onMouseDown={() => {onClose(); useEventStore.setState({error: null});}}>
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
        <div className="modal-backdrop" onMouseDown={() => {onClose(); useUsersStore.setState({error: null});}}>
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
                        onClick={() => {onClose(); useUsersStore.setState({error: null});}}
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
                                    <option value="Admin">{UserRoles.Admin}</option>
                                    <option value="Owner">{UserRoles.Owner}</option>
                                </select>
                            </div>
                            
                            {shouldShowClass && (
                                <div className="field">
                                    <label className="field__label" htmlFor="userClass">
                                        Класс
                                    </label>

                                    <select
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
                                    </select>
                                </div>
                            )}

                        </div>
                        
                    </div>

                    <div className="modal__footer">
                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => {onClose(); useUsersStore.setState({error: null});}}
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