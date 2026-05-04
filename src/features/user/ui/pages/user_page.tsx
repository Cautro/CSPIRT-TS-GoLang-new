import { type ReactNode, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { UserRoles } from "../../../../shared/entities/user/types/user_types";
import { NoteCard } from "../../../../shared/ui/note_card";
import { ComplaintCard } from "../../../../shared/ui/complaint_card";

import { useAuthStore } from "../../../auth/store/auth_store";
import { useUserStore } from "../../store/user_store";

import { noteAddDto } from "../../../../shared/entities/notes/api/notes_api";
import { complaintAddDto } from "../../../../shared/entities/complaints/api/complaints_api";
import { ratingChangeDTO } from "../../../../shared/entities/rating/api/rating_api";

export function UserPage() {
    const navigate = useNavigate();
    const { id } = useParams<{ id: string }>();
    const userId = Number(id ?? 0); 

    const currentUser = useAuthStore((state) => state.user?.User);

    const status = useUserStore((state) => state.status);
    const error = useUserStore((state) => state.error);
    const user = useUserStore((state) => state.user);

    const getUser = useUserStore((state) => state.getUser);
    const addNote = useUserStore((state) => state.addNote);
    const deleteNote = useUserStore((state) => state.deleteNote);
    const addComplaint = useUserStore((state) => state.addComplaint);
    const deleteComplaint = useUserStore((state) => state.deleteComplaint);
    const changeRating = useUserStore((state) => state.changeRating);
    const deleteUser = useUserStore((state) => state.deleteUser);

    const [noteText, setNoteText] = useState("");
    const [complaintText, setComplaintText] = useState("");
    const [ratingReason, setRatingReason] = useState("");
    const [ratingValue, setRatingValue] = useState("");
    const [formError, setFormError] = useState<string | null>(null);
    const [isDeleteUserModalOpen, setIsDeleteUserModalOpen] = useState(false);

    const isLoading = status === "loading";

    useEffect(() => {
        if (!id) {
            return;
        }

        void getUser(userId);
    }, [id, getUser]);

    async function refreshUser() {
        if (!id) {
            return;
        }

        await getUser(userId);
    }

    async function handleNoteAdd() {
        setFormError(null);

        if (!user || !currentUser) {
            setFormError("Не удалось определить пользователя");
            return;
        }

        const dto = {
            TargetID: user.User.Id,
            Content: noteText.trim(),
            AuthorID: currentUser.Id,
            CreatedAt: new Date().toISOString(),
            AuthorName: `${currentUser.Name} ${currentUser.LastName}`,
            TargetName: `${user.User.Name} ${user.User.LastName}`,
        };

        const parsed = noteAddDto.safeParse(dto);

        if (!parsed.success) {
            setFormError("Проверьте текст заметки");
            return;
        }

        await addNote(parsed.data);
        setNoteText("");
        await refreshUser();
    }

    async function handleComplaintAdd() {
        setFormError(null);

        if (!user || !currentUser) {
            setFormError("Не удалось определить пользователя");
            return;
        }

        const dto = {
            TargetID: user.User.Id,
            Content: complaintText.trim(),
            AuthorID: currentUser.Id,
            CreatedAt: new Date().toISOString(),
            AuthorName: `${currentUser.Name} ${currentUser.LastName}`,
            TargetName: `${user.User.Name} ${user.User.LastName}`,
        };

        const parsed = complaintAddDto.safeParse(dto);

        if (!parsed.success) {
            setFormError("Проверьте текст жалобы");
            return;
        }

        await addComplaint(parsed.data);
        setComplaintText("");
        await refreshUser();
    }

    async function handleChangeRating() {
        setFormError(null);

        if (!user) {
            setFormError("Пользователь не загружен");
            return;
        }

        const ratingNumber = Number(ratingValue);

        const dto = {
            rating: ratingNumber,
            target_login: user.User.Login,
            reason: ratingReason.trim(),
        };

        const parsed = ratingChangeDTO.safeParse(dto);

        if (!parsed.success) {
            setFormError("Проверьте значение рейтинга и причину изменения");
            return;
        }

        await changeRating(parsed.data);
        setRatingValue("");
        setRatingReason("");
        await refreshUser();
    }

    if (!user) {
        return (
            <main className="main">
                <section className="page">
                    {isLoading && (
                        <div className="profile-loading">
                            <div className="skeleton" style={{ height: 120 }} />
                            <div className="skeleton" style={{ height: 240 }} />
                        </div>
                    )}

                    {!isLoading && error && (
                        <div className="alert alert--danger">{error}</div>
                    )}

                    {!isLoading && !error && (
                        <div className="empty-state">
                            <h2 className="empty-state__title">Профиль не загружен</h2>
                            <p className="empty-state__text">
                                Данные пользователя отсутствуют или сессия недействительна.
                            </p>
                        </div>
                    )}
                </section>
            </main>
        );
    }

    const targetRole = user.User.Role;
    const currentRole = currentUser?.Role;
    const isYou = user.User.Id === currentUser?.Id;

    const isStudentLikeUser = targetRole === "User" || targetRole === "Helper";
    const canManageNotes =
        currentRole === "Helper" || currentRole === "Owner" || currentRole === "Admin";
    const canManageRating = currentRole === "Owner" || currentRole === "Admin";

    const notes = user.Notes ?? [];
    const complaints = user.Complaints ?? [];

    const fullName = `${user.User.Name ?? ""} ${user.User.LastName ?? ""}`.trim();
    const initials =
        `${user.User.Name?.[0] ?? ""}${user.User.LastName?.[0] ?? ""}` || "?";

    const rating = user.User.Rating ?? 0;
    const ratingPercent = Math.min(Math.max((rating / 5000) * 100, 0), 100);

    const ratingLevel =
        rating < 1500 ? "low" : rating < 3500 ? "medium" : "high";

    return (
        <main className="main">
            <section className="page user-page">
                <div className="user-hero">
                    <div className="user-hero__main">
                        <div className="profile-avatar">{initials}</div>

                        <div className="user-hero__content">
                            <h1 className="profile-hero__name">{fullName || "Без имени"}</h1>

                            <div className="profile-hero__meta">
                <span className="badge badge--info">
                  {UserRoles[targetRole] ?? targetRole}
                </span>

                                {isStudentLikeUser && (
                                    <span className="badge badge--neutral">
                    Класс {user.User.Class}
                  </span>
                                )}

                                <span className="profile-login">@{user.User.Login}</span>
                            </div>
                        </div>
                    </div>

                    <div className="profile-actions">

                        {currentRole === "Owner" && (
                            <button
                                className="btn btn--danger"
                                type="button"
                                onClick={async () => {
                                    if (id !== null) {
                                        setIsDeleteUserModalOpen(true);
                                    }
                                }}
                            >
                                Удалить пользователя
                            </button>
                        )}
                        
                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => navigate(-1)}
                        >
                            Назад
                        </button>
                    </div>
                </div>

                {isLoading && <div className="profile-progress" />}

                {error && <div className="alert alert--danger">{error}</div>}
                {formError && <div className="alert alert--danger">{formError}</div>}

                <div className={isStudentLikeUser ? "user-main-grid" : "user-main-grid user-main-grid--single"}>
                    <section className="card card--padded">
                        <div className="section-head">
                            <h2 className="section-title">Основная информация</h2>
                            <p className="section-description">
                                Базовые данные пользователя системы.
                            </p>
                        </div>

                        <div className="info-list">
                            <InfoRow label="ID пользователя" value={user.User.Id} />
                            <InfoRow label="Имя" value={user.User.Name} />
                            <InfoRow label="Фамилия" value={user.User.LastName} />
                            <InfoRow label="Полное имя" value={fullName || "Не указано"} />
                            <InfoRow label="Логин" value={user.User.Login} />

                            {isStudentLikeUser && (
                                <InfoRow label="Класс" value={user.User.Class} />
                            )}

                            <InfoRow
                                label="Роль"
                                value={UserRoles[targetRole] ?? targetRole}
                            />
                        </div>
                    </section>

                    {isStudentLikeUser && (
                        <section className="card card--padded user-rating-card">
                            <div className="section-head">
                                <h2 className="section-title">Рейтинг</h2>
                                <p className="section-description">
                                    Текущий социальный рейтинг пользователя.
                                </p>
                            </div>

                            <div className="user-rating-summary">
                                <div>
                                    <div className={`profile-rating-value profile-rating-value--${ratingLevel}`}>
                                        {rating}
                                    </div>
                                    <div className="text-muted">из 5000</div>
                                </div>

                                <span className={`badge badge--${ratingLevel}`}>
                  {Math.round(ratingPercent)}%
                </span>
                            </div>

                            <div className="rating">
                                <div className="rating__bar">
                                    <div
                                        className={`rating__fill rating__fill--${ratingLevel}`}
                                        style={{ width: `${ratingPercent}%` }}
                                    />
                                </div>
                            </div>

                            {canManageRating && (
                                <div className="user-action-form user-action-form--compact">
                                    <div className="field">
                                        <label className="field__label" htmlFor="ratingValue">
                                            Новое значение рейтинга
                                        </label>

                                        <input
                                            className="input"
                                            id="ratingValue"
                                            placeholder="Например: 4200"
                                            type="number"
                                            min={0}
                                            max={5000}
                                            value={ratingValue}
                                            onChange={(event) => setRatingValue(event.target.value)}
                                        />
                                    </div>

                                    <div className="field">
                                        <label className="field__label" htmlFor="ratingReason">
                                            Причина изменения
                                        </label>

                                        <textarea
                                            id="ratingReason"
                                            className="textarea"
                                            placeholder="Укажите причину изменения рейтинга"
                                            maxLength={500}
                                            value={ratingReason}
                                            onChange={(event) => setRatingReason(event.target.value)}
                                        />
                                    </div>

                                    <div className="user-action-form__footer">
                                        <button
                                            className="btn btn--primary"
                                            type="button"
                                            disabled={!ratingValue || !ratingReason.trim() || isLoading}
                                            onClick={() => void handleChangeRating()}
                                        >
                                            Изменить рейтинг
                                        </button>
                                    </div>
                                </div>
                            )}
                        </section>
                    )}
                </div>

                {isStudentLikeUser && (
                    <div className="user-content-grid">
                        <section className="card card--padded user-section-card">
                            <div className="section-head section-head--row">
                                <div>
                                    <h2 className="section-title">Жалобы</h2>
                                    <p className="section-description">
                                        Жалобы, связанные с текущим пользователем.
                                    </p>
                                </div>

                                <span
                                    className={
                                        complaints.length > 0
                                            ? "badge badge--danger"
                                            : "badge badge--neutral"
                                    }
                                >
                  {complaints.length}
                </span>
                            </div>

                            {!isYou && (<div className="user-action-form">
                                <div className="field">
                                    <label className="field__label" htmlFor="complaintText">
                                        Новая жалоба
                                    </label>

                                    <textarea
                                        id="complaintText"
                                        className="textarea"
                                        placeholder="Введите текст жалобы на пользователя..."
                                        maxLength={500}
                                        value={complaintText}
                                        onChange={(event) => setComplaintText(event.target.value)}
                                    />
                                </div>

                                <div className="user-action-form__footer">
                                    <button
                                        className="btn btn--danger"
                                        type="button"
                                        disabled={!complaintText.trim() || isLoading}
                                        onClick={() => void handleComplaintAdd()}
                                    >
                                        Отправить жалобу
                                    </button>
                                </div>
                            </div>)}

                            {complaints.length > 0 ? (
                                <div className="feed">
                                    {complaints.map((item) => (
                                        <ComplaintCard
                                            key={item.ID}
                                            item={item}
                                            onDelete={async () => {
                                                await deleteComplaint(item.ID);
                                                await refreshUser();
                                            }}
                                            role={currentRole ?? "User"}
                                        />
                                    ))}
                                </div>
                            ) : (
                                <div className="empty-inline">Жалоб нет</div>
                            )}
                        </section>

                        {canManageNotes && (
                            <section className="card card--padded user-section-card">
                                <div className="section-head section-head--row">
                                    <div>
                                        <h2 className="section-title">Заметки</h2>
                                        <p className="section-description">
                                            Поведенческие заметки, оставленные ответственными пользователями.
                                        </p>
                                    </div>

                                    <span className="badge badge--neutral">{notes.length}</span>
                                </div>

                                {!isYou && (<div className="user-action-form">
                                    <div className="field">
                                        <label className="field__label" htmlFor="noteText">
                                            Новая заметка
                                        </label>

                                        <textarea
                                            id="noteText"
                                            className="textarea"
                                            placeholder="Введите заметку о поведении пользователя..."
                                            maxLength={500}
                                            value={noteText}
                                            onChange={(event) => setNoteText(event.target.value)}
                                        />
                                    </div>

                                    <div className="user-action-form__footer">
                                        <button
                                            className="btn btn--primary"
                                            type="button"
                                            disabled={!noteText.trim() || isLoading}
                                            onClick={() => void handleNoteAdd()}
                                        >
                                            Добавить заметку
                                        </button>
                                    </div>
                                </div>)}

                                {notes.length > 0 ? (
                                    <div className="feed">
                                        {notes.map((note) => (
                                            <NoteCard
                                                key={note.ID}
                                                item={note}
                                                onDelete={async () => {
                                                    await deleteNote(note.ID);
                                                    await refreshUser();
                                                }}
                                                role={currentRole}
                                            />
                                        ))}
                                    </div>
                                ) : (
                                    <div className="empty-inline">Заметок нет</div>
                                )}
                            </section>
                        )}
                    </div>
                )}

                {isDeleteUserModalOpen && (
                    <div className="modal-backdrop" onMouseDown={() => setIsDeleteUserModalOpen(false)}>
                        <section
                            className="modal modal--confirm"
                            role="dialog"
                            aria-modal="true"
                            aria-labelledby="delete-event-title"
                            onMouseDown={(event) => event.stopPropagation()}
                        >
                            <div className="modal__header">
                                <div>
                                    <h2 className="modal__title" id="delete-event-title">
                                        Удалить пользователя?
                                    </h2>

                                    <p className="modal__description">
                                        Это действие удалит пользователя '{user.User.Name} {user.User.LastName}'. Отменить удаление будет нельзя.
                                    </p>
                                </div>

                                <button
                                    className="modal__close"
                                    type="button"
                                    onClick={() => setIsDeleteUserModalOpen(false)}
                                    aria-label="Закрыть окно подтверждения"
                                    disabled={isLoading}
                                >
                                    ×
                                </button>
                            </div>

                            <div className="modal__footer">
                                <button
                                    className="btn btn--secondary"
                                    type="button"
                                    onClick={() => setIsDeleteUserModalOpen(false)}
                                    disabled={isLoading}
                                >
                                    Отмена
                                </button>

                                <button
                                    className="btn btn--danger"
                                    type="button"
                                    disabled={isLoading}
                                    onClick={async () => {
                                        if (id !== null) {
                                            await deleteUser(userId);
                                            setIsDeleteUserModalOpen(false);
                                            navigate(-1);
                                        }
                                    }}
                                >
                                    {isLoading ? "Удаление..." : "Удалить"}
                                </button>
                            </div>
                        </section>
                    </div>
                )}
                
            </section>
        </main>
    );
}

interface InfoRowProps {
    label: string;
    value: ReactNode;
}

function InfoRow({ label, value }: InfoRowProps) {
    return (
        <div className="info-row">
            <span className="info-row__label">{label}</span>
            <span className="info-row__value">{value || "Не указано"}</span>
        </div>
    );
}