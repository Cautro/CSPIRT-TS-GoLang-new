import {type ReactNode} from "react";

import { useAuthStore } from "../../../features/auth/store/auth_store.ts";
import { UserRoles } from "../../../shared/entities/user/types/user_types.ts";
import {TeacherScheduleWidget} from "../../../features/schedule/ui/components/teacher_schedule_widget.tsx";
import {type BurgerDrawerMenuItem} from "../../../shared/ui/other/burger_menu.tsx";
import {PageHeader} from "../../../shared/ui/other/page_header.tsx";
import {ComplaintsSection} from "../../../features/complaints/ui/components/complaints_section.tsx";
import {NotesSection} from "../../../features/notes/ui/components/notes_section.tsx";
import {RatingSection} from "../../../features/rating/ui/components/rating_section.tsx";
import {useLogout} from "../../../features/auth/hooks/use_logout.ts";

export function ProfilePage() {
    const profile = useAuthStore((state) => state.user);
    const logout = useLogout()
    const status = useAuthStore((state) => state.status);
    const error = useAuthStore((state) => state.error);

    const isLoading = status === "loading";
    
    const menuItems: BurgerDrawerMenuItem[] = [
        {
            label: "Выйти",
            onClick: async () => {
                await logout.mutateAsync()
            },
        }
    ]

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

    const user = profile.User;
    const notes = profile.Notes ?? [];
    const complaints = profile.Complaints ?? [];

    const fullName = `${user.Name ?? ""} ${user.LastName ?? ""}`.trim();
    const normalizedRole = user.Role.toLowerCase();

    const isStudentLikeUser = normalizedRole === "user" || normalizedRole === "helper";
    
    return (
        <main className="main">
            <section className="page user-page">
                
                <PageHeader
                    title={`${fullName}`}
                    meta={
                        <>
                            <span className="badge badge--info">
                                    {UserRoles[user.Role] ?? user.Role}
                                </span>

                            {isStudentLikeUser && (
                                <span className="badge badge--neutral">
                                        Класс {user.Class}
                                    </span>
                            )}
                            
                            <span className="profile-login">@{user.Login}</span>
                        </>
                    }
                    menuItems={menuItems}
                    hasBackButton={true}
                />

                {isLoading && <div className="profile-progress" />}

                {error && <div className="alert alert--danger">{error}</div>}

                <div
                    className={
                        isStudentLikeUser
                            ? "user-main-grid"
                            : "user-main-grid user-main-grid--single"
                    }
                >
                    <section className="card card--padded">
                        <div className="section-head">
                            <h2 className="section-title">Основная информация</h2>
                            <p className="section-description">
                                Базовые данные текущего пользователя системы.
                            </p>
                        </div>

                        <div className="info-list">
                            <InfoRow label="Имя" value={user.Name} />
                            <InfoRow label="Фамилия" value={user.LastName} />
                            <InfoRow label="Полное имя" value={fullName || "Не указано"} />
                            <InfoRow label="Логин" value={user.Login} />

                            {isStudentLikeUser && (
                                <InfoRow label="Класс" value={user.Class} />
                            )}

                            <InfoRow
                                label="Роль"
                                value={UserRoles[user.Role] ?? user.Role}
                            />
                        </div>
                    </section>

                    {(normalizedRole === "admin" || normalizedRole === "owner") && (
                        <section className="card card--padded">
                            <div className="section-head">
                                <h2 className="section-title">Расписание учителя</h2>
                                <p className="section-description">
                                    Подробные данные о вашем расписании как учителя на текущую неделю.
                                </p>
                            </div>

                            <TeacherScheduleWidget name={`${user.Name} ${user.LastName}`} id={user.Id}/>

                        </section>
                    )}

                    {isStudentLikeUser && (
                        <RatingSection user={user} />
                    )}
                </div>

                {isStudentLikeUser && (
                    <div className="user-content-grid">
                        <ComplaintsSection
                            isProfile={true}
                            complaints={complaints}
                            user={user}
                            isYou={true}
                            currentUser={user}
                        />

                       <NotesSection
                         isProfile={true}
                         notes={notes}    
                         user={user}
                         isYou={true}
                         currentUser={user}
                       />
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