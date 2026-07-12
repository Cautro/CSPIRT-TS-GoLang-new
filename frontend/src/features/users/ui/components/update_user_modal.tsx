import {type ChangeEvent, type FormEvent, useEffect, useState} from "react";
import {type UserRole, UserRoles, type UserType} from "../../../../shared/entities/user/types/user_types.ts";
import {useClasses} from "../../../class/hooks/use_classes.ts";
import {useUpdateUser} from "../../hooks/use_update_user.ts";
import type {updateUserValues} from "../../models/update_user_usecase.ts";
import {compressImage} from "../../../../core/image/image_compress.ts";

interface AddUserModalProps {
    isOpen: boolean;
    onClose: () => void;
    onAddUser: () => void;
    classId?: number | null;
    user: UserType;
}

export function UpdateUserModal({isOpen, onClose, onAddUser, classId = null, user}: AddUserModalProps) {
    const classes = useClasses().data;
    const {mutateAsync, error} = useUpdateUser()
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [selectedRole, setSelectedRole] = useState<UserRole>("User");

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setSelectedRole(user.Role);
    }, [user.Role]);
    const normalizedSelectedRole = selectedRole.toLowerCase();
    const shouldShowClass = normalizedSelectedRole === "user" || normalizedSelectedRole === "helper";
    const [selectedImage, setSelectedImage] = useState(user.Avatar.String)

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

    const handleChangeImage = async (e: ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const compressed = await compressImage(file, 512, 512, 0.75);
        setSelectedImage(compressed);
    };

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        setIsSubmitting(true)

        const formData = new FormData(event.currentTarget);

        const selectedClassId = classId ?? (shouldShowClass ? Number(formData.get("classId")) : 0);

        const selectedClassName = classes?.find(c => c.Id === selectedClassId)?.Name ?? "";

        const form: updateUserValues = {
            rating: user.Rating,
            id: user.Id,
            avatar: selectedImage || user.Avatar.String,
            name: String(formData.get("name") ?? "").trim(),
            lastname: String(formData.get("lastName") ?? "").trim(),
            classId: selectedClassId,
            login: String(formData.get("login") ?? "").trim(),
            role: selectedRole,
            className: selectedClassName,
        };
        
        console.log(form, classes)

        await mutateAsync(form);
        onAddUser();
        setIsSubmitting(false);
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
                            Изменение пользователя
                        </h2>

                        <p className="modal__description">
                            Заполните новые данные о пользователе
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
                                <label className="field__label">
                                    Иконка пользователя
                                </label>
                                <div className="avatar-picker">
                                    <label htmlFor="avatar-upload" className="avatar-picker__label">
                                        {selectedImage ? (
                                            <img
                                                src={selectedImage}
                                                alt="preview"
                                                className="avatar-picker__preview"
                                            />
                                        ) : (
                                            <div className="avatar-picker__placeholder">
                                                +
                                            </div>
                                        )}
                                    </label>

                                    <input
                                        id="avatar-upload"
                                        type="file"
                                        accept="image/*"
                                        onChange={handleChangeImage}
                                        hidden
                                    />

                                    <button
                                        type="button"
                                        className="btn btn--secondary"
                                        onClick={() =>
                                            document.getElementById("avatar-upload")?.click()
                                        }
                                    >
                                        Выбрать изображение
                                    </button>
                                </div>
                            </div>

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
                                    defaultValue={user.Name}
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
                                    defaultValue={user.LastName}
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
                                    defaultValue={user.Login}
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
                                        defaultValue={user.ClassID}
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
                                            <option key={item.Id} value={item.Id}>
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
                            {isSubmitting ? "Изменение..." : "Изменить пользователя"}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    );
}