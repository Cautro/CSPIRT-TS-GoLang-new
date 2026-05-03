import { type FormEvent, useEffect, useState } from "react";
import type { ClassType } from "../../../../shared/entities/class/types/class_types";
import {type UserRole, UserRoles} from "../../../../shared/entities/user/types/user_types";
import { addUserDto, type addUserType } from "../../../../shared/entities/user/api/user_api";

interface AddUserModalProps {
    isOpen: boolean;
    onClose: () => void;
    classes: ClassType[];
    onAddUser: (dto: addUserType) => Promise<void>;
}

export function AddUserModal({
                                 isOpen,
                                 onClose,
                                 classes,
                                 onAddUser,
                             }: AddUserModalProps) {
    const [formError, setFormError] = useState<string | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [selectedRole, setSelectedRole] = useState<UserRole>("User");
    const shouldShowClass = selectedRole === "User" || selectedRole === "Helper";

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
        const lastName = String(formData.get("lastName") ?? "").trim();
        const login = String(formData.get("login") ?? "").trim();
        const password = String(formData.get("password") ?? "").trim();
        const classIdRaw = String(formData.get("classId") ?? "");
        const role = String(formData.get("role") ?? "User").trim() as UserRole;
        
        const dto = {
            Name: name,
            LastName: lastName,
            FullName: [
                {
                    Name: name,
                    LastName: lastName,
                },
            ],
            Password: password,
            ClassID: shouldShowClass ? Number(classIdRaw) : 0,
            Login: login,
            Role: role,
        };

        const parsed = addUserDto.safeParse(dto);

        if (!parsed.success) {
            console.log(parsed.error.format());
            setFormError("Проверьте правильность заполнения полей");
            return;
        }

        try {
            setIsSubmitting(true);
            await onAddUser(parsed.data);
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
                            Добавление пользователя
                        </h2>

                        <p className="modal__description">
                            Заполните данные нового пользователя системы.
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
                            {isSubmitting ? "Создание..." : "Создать пользователя"}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    );
}