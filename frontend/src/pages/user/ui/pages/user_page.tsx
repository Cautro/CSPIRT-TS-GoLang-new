import { type ReactNode, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {UserRoles, type UserType} from "../../../../shared/entities/user/types/user_types.ts";
import { useAuthStore } from "../../../../features/auth/store/auth_store.ts";
import {ConfirmModal} from "../../../../shared/ui/modals/confirm_modal.tsx";
import {TeacherScheduleWidget} from "../../../../features/schedule/ui/components/teacher_schedule_widget.tsx";
import {type BurgerDrawerMenuItem} from "../../../../shared/ui/other/burger_menu.tsx";
import {PageHeader} from "../../../../shared/ui/other/page_header.tsx";
import {useUser} from "../../../../features/users/hooks/use_user.ts";
import {NotesSection} from "../../../../features/notes/ui/components/notes_section.tsx";
import {ComplaintsSection} from "../../../../features/complaints/ui/components/complaints_section.tsx";
import {RatingSection} from "../../../../features/rating/ui/components/rating_section.tsx";
import {useDeleteUser} from "../../../../features/users/hooks/use_delete_user.ts";
import {UpdateUserModal} from "../../../../features/users/ui/components/update_user_modal.tsx";

export function UserPage() {
    const navigate = useNavigate();
    const { id } = useParams<{ id: string }>();
    const userId = Number(id ?? 0); 

    const currentUser = useAuthStore((state) => state.user?.User);
    const getUser = useUser(userId)
    const user = getUser.data
    const deleteUser = useDeleteUser()
    
    const [formError, setFormError] = useState<string | null>(null);
    const [isDeleteUserModalOpen, setIsDeleteUserModalOpen] = useState(false);
    const [isUpdateUserModalOpen, setIsUpdateUserModalOpen] = useState(false);

    const error = getUser.error?.message ?? null;
    const isLoading = getUser.isLoading;

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
    const normalizedTargetRole = targetRole.toLowerCase();
    const normalizedCurrentRole = currentRole?.toLowerCase();
    const isYou = user.User.Id === currentUser?.Id;

    const isStudentLikeUser = normalizedTargetRole === "user" || normalizedTargetRole === "helper";
    const canManageNotes =
        normalizedCurrentRole === "helper" || normalizedCurrentRole === "owner" || normalizedCurrentRole === "admin";

    const notes = user.Notes ?? [];
    const complaints = user.Complaints ?? [];

    const fullName = `${user.User.Name ?? ""} ${user.User.LastName ?? ""}`.trim();
    
    const menuItems: BurgerDrawerMenuItem[] = [
        {
            label: "Удалить пользователя",
            hidden: (normalizedCurrentRole !== "owner"),
            onClick: () => setIsDeleteUserModalOpen(true)
        },
        {
            label: "Изменить пользователя",
            hidden: (normalizedCurrentRole !== "owner"),
            onClick: () => setIsUpdateUserModalOpen(true)
        }
    ]

    return (
        <main className="main">
            <section className="page user-page">
                
                <PageHeader 
                    title={fullName || "Без имени"}
                    meta={
                        <>
                            <span className="badge badge--info">
                                    {UserRoles[targetRole] ?? targetRole}
                                </span>

                            {isStudentLikeUser && (
                                <span className="badge badge--neutral">
                                        Класс {user.User.Class}
                                    </span>
                            )}

                            <span className="profile-login">@{user.User.Login}</span>
                        </>
                    }
                    hasBackButton={true}
                    menuItems={menuItems}
                    menuTitle={"Меню"}
                />

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
                            <InfoRow label="Имя" value={user.User.Name} />
                            <InfoRow label="Фамилия" value={user.User.LastName} />
                            <InfoRow label="Полное имя" value={fullName || "Не указано"} />
                            {normalizedCurrentRole === "owner" && (<InfoRow label="Логин" value={user.User.Login}/>)}

                            {isStudentLikeUser && (
                                <InfoRow label="Класс" value={user.User.Class} />
                            )}

                            <InfoRow
                                label="Роль"
                                value={UserRoles[targetRole] ?? targetRole}
                            />
                        </div>
                    </section>

                    {(normalizedTargetRole === "admin" || normalizedTargetRole === "owner") && (normalizedCurrentRole === "owner") && (
                        <section className="card card--padded">
                            <div className="section-head">
                                <h2 className="section-title">Расписание учителя</h2>
                                <p className="section-description">
                                    Подробные данные о расписании конкретного учителя на текущую неделю.
                                </p>
                            </div>
                            
                            <TeacherScheduleWidget name={`${user.User.Name} ${user.User.LastName}`} id={user.User.Id} />
                            
                        </section>
                    )}

                    {isStudentLikeUser && (
                        <RatingSection
                            user={user.User}
                            setFormError={setFormError}
                        />
                    )}
                </div>

                {isStudentLikeUser && (
                    <div className="user-content-grid">

                        <ComplaintsSection
                            user={user.User}
                            currentUser={currentUser as UserType}
                            complaints={complaints}
                            isYou={isYou}
                            setFormError={setFormError}
                        />

                        {canManageNotes && (
                            <NotesSection
                                notes={notes}
                                user={user.User}
                                currentUser={currentUser as UserType}
                                isYou={isYou}
                                setFormError={setFormError}
                            />
                        )}
                        
                    </div>
                )}

                {isUpdateUserModalOpen && (
                    <UpdateUserModal 
                        isOpen={isUpdateUserModalOpen}
                         onClose={() => setIsUpdateUserModalOpen(false)} 
                         onAddUser={() => setIsUpdateUserModalOpen(false)} 
                         user={user.User}
                    />
                )}

                {isDeleteUserModalOpen && (
                    <ConfirmModal
                        content={`Это действие удалит пользователя "${user.User.Name} ${user.User.LastName}". Отменить удаление будет нельзя.`}
                        onClose={() => setIsDeleteUserModalOpen(false)}
                        onConfirm={async () => {
                            if (id !== null) {
                                await deleteUser.mutateAsync({id: userId});
                                setIsDeleteUserModalOpen(false);
                                navigate(-1);
                            }
                        }}
                        isOpen={isDeleteUserModalOpen}
                        buttonContent={"Удалить"}
                        isDanger={true}
                    />
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