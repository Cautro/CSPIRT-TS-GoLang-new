import type { ReactNode } from "react";
import { useNavigate } from "react-router-dom";

import { useAuthStore } from "../../auth/store/auth_store";
import { UserRoles } from "../../../shared/entities/user/types/user_types";
import { truncateText } from "../../../core/security/security_limits";
import {NoteCard} from "../../../shared/ui/note_card.tsx";

export function ProfilePage() {
    const navigate = useNavigate();

    const profile = useAuthStore((state) => state.user);
    const getProfile = useAuthStore((state) => state.checkAuth);
    const logout = useAuthStore((state) => state.logout);
    const status = useAuthStore((state) => state.status);
    const error = useAuthStore((state) => state.error);

    const isLoading = status === "loading";

    function safeUnknownToText(value: unknown): string {
        if (typeof value === "string") {
            return truncateText(value, 500);
        }

        if (
            typeof value === "object" &&
            value !== null &&
            "Text" in value &&
            typeof (value as { Text?: unknown }).Text === "string"
        ) {
            return truncateText((value as { Text: string }).Text, 500);
        }

        return "Скрыто: неизвестный формат данных";
    }

    async function handleLogout() {
        await logout();
        navigate("/login", { replace: true });
    }

    if (!profile) {
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

    const notes = profile.Notes ?? [];
    const complaints = profile.Complaints ?? [];

    const fullName = `${profile.User.Name ?? ""} ${profile.User.LastName ?? ""}`.trim();
    const initials =
        `${profile.User.Name?.[0] ?? ""}${profile.User.LastName?.[0] ?? ""}` || "?";

    const rating = profile.User.Rating ?? 0;
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
                  {UserRoles[profile.User.Role] ?? profile.User.Role}
                </span>

                                <span className="badge badge--neutral">
                  Класс {profile.User.Class}
                </span>

                                <span className="profile-login">@{profile.User.Login}</span>
                            </div>
                        </div>
                    </div>

                    <div className="profile-actions">
                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => navigate("/", { replace: true })}
                        >
                            Главная
                        </button>

                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => void getProfile()}
                            disabled={isLoading}
                        >
                            Обновить
                        </button>

                        <button
                            className="btn btn--danger"
                            type="button"
                            onClick={() => void handleLogout()}
                            disabled={isLoading}
                        >
                            Выйти
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
                            <InfoRow label="ID пользователя" value={profile.User.Id} />
                            <InfoRow label="Имя" value={profile.User.Name} />
                            <InfoRow label="Фамилия" value={profile.User.LastName} />
                            <InfoRow label="Полное имя" value={fullName || "Не указано"} />
                            <InfoRow label="Логин" value={profile.User.Login} />
                            <InfoRow label="Класс" value={profile.User.Class} />
                            <InfoRow
                                label="Роль"
                                value={UserRoles[profile.User.Role] ?? profile.User.Role}
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
                    </section>
                </div>

                <div className="profile-grid profile-grid--equal">
                    <section className="card card--padded">
                        <div className="section-head section-head--row">
                            <div>
                                <h2 className="section-title">Заметки</h2>
                                <p className="section-description">
                                    Поведенческие заметки, оставленные ответственными пользователями.
                                </p>
                            </div>

                            <span className="badge badge--neutral">{notes.length}</span>
                        </div>

                        {notes.length > 0 ? (
                            <div className="feed">
                                {notes.map((note) => (
                                    <NoteCard item={note} role={"User"}/>
                                ))}
                            </div>
                        ) : (
                            <div className="empty-inline">Заметок нет</div>
                        )}
                    </section>

                    <section className="card card--padded">
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

                        {complaints.length > 0 ? (
                            <div className="feed">
                                {complaints.map((complaint, index) => (
                                    <div className="feed-item feed-item--warning" key={index}>
                                        <p className="feed-item__text">
                                            {safeUnknownToText(complaint)}
                                        </p>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div className="empty-inline">Жалоб нет</div>
                        )}
                    </section>
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