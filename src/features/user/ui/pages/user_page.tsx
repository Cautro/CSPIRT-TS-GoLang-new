import {type ReactNode, useEffect, useState} from "react";
import {useNavigate, useParams} from "react-router-dom";
import {UserRoles} from "../../../../shared/entities/user/types/user_types.ts";
import {NoteCard} from "../../../../shared/ui/note_card.tsx";
import {useAuthStore} from "../../../auth/store/auth_store.ts";
import {useUserStore} from "../../store/user_store.ts";
import {noteAddDto,} from "../../../../shared/entities/notes/api/notes_api.ts";
import {complaintAddDto} from "../../../../shared/entities/complaints/api/complaints_api.ts";
import {ComplaintCard} from "../../../../shared/ui/complaint_card.tsx";
import {ratingChangeDTO} from "../../../../shared/entities/rating/api/rating_api.ts";

export function UserPage() {
    const navigate = useNavigate();
    const { id } = useParams<{ id: string }>(); 
    const role = useAuthStore.getState().user?.User.Role;
    const name = useAuthStore.getState().user?.User.Name
    const lastName = useAuthStore.getState().user?.User.LastName
    
    const status = useUserStore((state) => state.status);
    const error = useUserStore((state) => state.error);
    const user = useUserStore((state) => state.user);
    const getUser = useUserStore((state) => state.getUser);
    const addNote = useUserStore((state) => state.addNote);
    const deleteNote = useUserStore((state) => state.deleteNote);
    const addComplaint = useUserStore((state) => state.addComplaint);
    const deleteComplaint = useUserStore((state) => state.deleteComplaint);
    const changeRating = useUserStore((state) => state.changeRating);
    
    const [noteText, setNoteText] = useState<string>("");
    const [complaintText, setComplaintText] = useState<string>("");
    const [changeRatingText, setChangeRatingText] = useState<string>("");
    const [changeRatingValue, setChangeRatingValue] = useState<number>();
    
    useEffect(() => {
        if (!id) {
            return;
        }
        void getUser(id);
    }, [id, getUser]);
    
    const isLoading = status === "loading";

    function handleNoteAdd() {
        if (!user) {
            return;
        }

        const dto = {
            TargetID: user.User.Id,
            Content: noteText.trim(),
            AuthorID: (useAuthStore.getState().user?.User.Id),
            CreatedAt: (new Date().toISOString()),
            AuthorName: `${name} ${lastName}`,
            TargetName: `${user.User.Name} ${user.User.LastName}`,
        };

        const parsed = noteAddDto.safeParse(dto);

        if (!parsed.success) {
            return;
        }

        void addNote(parsed.data);
        setNoteText("");
    }

    function handleComplaintAdd() {
        if (!user) {
            return;
        }

        const dto = {
            TargetID: user.User.Id,
            Content: complaintText.trim(),
            AuthorID: (useAuthStore.getState().user?.User.Id),
            CreatedAt: (new Date().toISOString()),
            AuthorName: `${name} ${lastName}`,
            TargetName: `${user.User.Name} ${user.User.LastName}`,
        };

        const parsed = complaintAddDto.safeParse(dto);

        if (!parsed.success) {
            return;
        }

        void addComplaint(parsed.data);
        setComplaintText("");
    }

    function handleChangeRating() {
        if (!user) {
            return;
        }

        const dto = {
            rating: changeRatingValue,
            target_login: user.User.Login,
            reason: changeRatingText,
        };
        

        const parsed = ratingChangeDTO.safeParse(dto);

        if (!parsed.success) {
            return;
        }
        void changeRating(parsed.data);
        setChangeRatingValue(0);
        setChangeRatingText("");
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

    const notes = user.Notes ?? [];
    const complaints = user.Complaints ?? [];

    const fullName = `${user.User.Name ?? ""} ${user.User.LastName ?? ""}`.trim();
    const initials =
        `${user.User.Name?.[0] ?? ""}${user.User.LastName?.[0] ?? ""}` || "?";

    const rating = user.User.Rating ?? 0;
    const ratingPercent = Math.min(Math.max((rating / 5000) * 100, 0), 100);

    return (
        <main className="main">
            <section className="page profile-page">
                <div className="profile-hero">
                    <div className="profile-hero__main">
                        <div className="profile-avatar">{initials}</div>

                        <div className="profile-hero__info">
                            <h1 className="profile-hero__name">{fullName || "Без имени"}</h1>

                            <div className="profile-hero__meta">
                <span className="badge badge--info">
                  {UserRoles[user.User.Role] ?? user.User.Role}
                </span>

                                <span className="badge badge--neutral">
                  Класс {user.User.Class}
                </span>

                                <span className="profile-login">@{user.User.Login}</span>
                            </div>
                        </div>
                    </div>

                    <div className="profile-actions">
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

                {error && <div className="alert alert--danger mb-4">{error}</div>}

                <div className="profile-grid">
                    <section className="card card--padded">
                        <div className="section-head">
                            <h2 className="section-title">Основная информация</h2>
                            <p className="section-description">
                                Базовые данные текущего пользователя системы.
                            </p>
                        </div>

                        <div className="info-list">
                            <InfoRow label="ID пользователя" value={user.User.Id} />
                            <InfoRow label="Имя" value={user.User.Name} />
                            <InfoRow label="Фамилия" value={user.User.LastName} />
                            <InfoRow label="Полное имя" value={fullName || "Не указано"} />
                            <InfoRow label="Логин" value={user.User.Login} />
                            <InfoRow label="Класс" value={user.User.Class} />
                            <InfoRow
                                label="Роль"
                                value={UserRoles[user.User.Role] ?? user.User.Role}
                            />
                        </div>
                    </section>

                    <section className="card card--padded profile-rating-card">
                        <div className="section-head">
                            <h2 className="section-title">Рейтинг</h2>
                            <p className="section-description">
                                Текущий социальный рейтинг пользователя.
                            </p>
                        </div>

                        <div className="profile-rating-value">{rating}</div>

                        <div className="rating">
                            <div className="rating__top">
                                <span className="rating__value">{rating} / 5000</span>
                                <span className="rating__value">{Math.round(ratingPercent)}%</span>
                            </div>

                            <div className="rating__bar">
                                <div
                                    className="rating__fill rating__fill--high"
                                    style={{ width: `${ratingPercent}%` }}
                                />
                            </div>
                        </div>

                        {(role === "Owner" || role === "Admin") && (
                            <div className="user-action-form">
                                <div className="field">
                                    <label className="field__label" htmlFor="noteText">
                                        Изменить рейтинг
                                    </label>

                                    <input
                                        className="input"
                                        id="noteText"
                                        placeholder="Измените рейтинг"
                                        type={"number"}
                                        onChange={(e) => setChangeRatingValue(parseInt(e.target.value))}
                                        value={changeRatingValue}
                                    />

                                    <textarea
                                        id="noteText"
                                        className="textarea"
                                        placeholder="Напишите причину изменения рейтинга"
                                        maxLength={500}
                                        onChange={(e) => setChangeRatingText(e.target.value)}
                                        value={changeRatingText}
                                    />
                                </div>

                                <div className="user-action-form__footer">
                                    <button className="btn btn--primary" type="button" onClick={() => {
                                        handleChangeRating();
                                        if (!id) {
                                            return
                                        }
                                        getUser(id);
                                    }}>
                                        Изменить рейтинг
                                    </button>
                                </div>
                            </div>
                        )}
                        
                    </section>
                </div>

                <div className="profile-grid profile-grid--equal">
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

                        <div className="user-action-form">
                            <div className="field">
                                <label className="field__label" htmlFor="complaintText">
                                    Новая жалоба
                                </label>

                                <textarea
                                    id="complaintText"
                                    className="textarea"
                                    placeholder="Введите текст жалобы на пользователя..."
                                    maxLength={500}
                                    onChange={(e) => setComplaintText(e.target.value)}
                                    value={complaintText}
                                />
                            </div>

                            <div className="user-action-form__footer">
                                <button className="btn btn--danger" type="button" onClick={() => {
                                    handleComplaintAdd();
                                    if (!id) {
                                        return
                                    }
                                    getUser(id);
                                }}>
                                    Отправить жалобу
                                </button>
                            </div>
                        </div>

                        {complaints.length > 0 ? (
                            <div className="feed">
                                {complaints.map((item, index) => (
                                    <ComplaintCard key={index} item={item} role={role} onDelete={() => {
                                        deleteComplaint(item.ID.toString());
                                        if (!id) {
                                            return
                                        }
                                        getUser(id);
                                    }} />
                                ))}
                            </div>
                        ) : (
                            <div className="empty-inline">Жалоб нет</div>
                        )}
                    </section>

                    {(role === "Helper" || role === "Owner" || role === "Admin") && (
                        <section className="card card--padded user-section-card">
                            <div className="section-head section-head--row">
                                <div>
                                    <h2 className="section-title">Заметки</h2>
                                    <p className="section-description">
                                        Поведенческие заметки, оставленные старостами.
                                    </p>
                                </div>

                                <span className="badge badge--neutral">{notes.length}</span>
                            </div>

                            <div className="user-action-form">
                                <div className="field">
                                    <label className="field__label" htmlFor="noteText">
                                        Новая заметка
                                    </label>

                                    <textarea
                                        id="noteText"
                                        className="textarea"
                                        placeholder="Введите заметку о поведении пользователя..."
                                        maxLength={500}
                                        onChange={(e) => setNoteText(e.currentTarget.value)}
                                        value={noteText}
                                    />
                                </div>

                                <div className="user-action-form__footer">
                                    <button className="btn btn--primary" type="button" onClick={() => {
                                        handleNoteAdd();
                                        if (!id) {
                                            return
                                        }
                                        getUser(id);
                                    }}>
                                        Добавить заметку
                                    </button>
                                </div>
                            </div>

                            {notes.length > 0 ? (
                                <div className="feed">
                                    {notes.map((note, index) => (
                                        <NoteCard key={index} item={note} role={role} onDelete={() => {
                                            deleteNote(note.ID.toString());
                                            if (!id) {
                                                return
                                            }
                                            getUser(id);
                                        }} />
                                    ))}
                                </div>
                            ) : (
                                <div className="empty-inline">Заметок нет</div>
                            )}
                        </section>
                    )}
                </div>
                
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