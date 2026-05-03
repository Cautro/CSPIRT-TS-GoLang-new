import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import { useAuthStore } from "../../../auth/store/auth_store";
import { useDashboardStore } from "../../store/dashboard_store";
import { ClassCard } from "../../../../shared/ui/class_card";
import { AddUserModal } from "../components/add_user_modal.tsx";
import {StaffCard} from "../../../../shared/ui/staff_card.tsx";
import {AddClassModal} from "../components/add_class_modal.tsx";
import {EventCard} from "../../../../shared/ui/event_card.tsx";
import {AddEventModal} from "../components/add_event_modal.tsx";

type Lists = "classes" | "events" | "staff";

export function DashboardPage() {
    const navigate = useNavigate();

    const role = useAuthStore((state) => state.user?.User.Role);
    const status = useDashboardStore((state) => state.status);
    const error = useDashboardStore((state) => state.error);
    const classes = useDashboardStore((state) => state.classes);
    const staff = useDashboardStore((state) => state.staff);
    const events = useDashboardStore((state) => state.events);
    const getStaff = useDashboardStore((state) => state.getStaff);
    const getClasses = useDashboardStore((state) => state.getClasses);
    const addUser = useDashboardStore((state) => state.addUser);
    const addClass = useDashboardStore((state) => state.addClass);
    const getEvents = useDashboardStore((state) => state.getEvents);
    const addEvent = useDashboardStore((state) => state.addEvent);

    const isLoading = status === "loading";

    const [selectedList, setSelectedList] = useState<Lists>("classes");
    const [isAddUserModalOpen, setIsAddUserModalOpen] = useState(false);
    const [isAddClassModalOpen, setIsAddClassModalOpen] = useState(false);
    const [isAddEventModalOpen, setIsAddEventModalOpen] = useState(false);

    useEffect(() => {
        void getClasses();
    }, [getClasses]);

    if (!role) {
        return null;
    }

    return (
        <main className="main">
            <section className="page">
                <div className="profile-hero">
                    <div className="info-row">
                        <p className="info-row__label">
                            Панель просмотра
                        </p>
                    </div>

                    <div className="btn-group">
                        {selectedList === "classes" && role === "Owner" && (
                            <div className="btn-group">
                                <button
                                    className="btn btn--primary"
                                    type="button"
                                    onClick={() => setIsAddUserModalOpen(true)}
                                >
                                    Добавить пользователя
                                </button>

                                <button
                                    className="btn btn--primary"
                                    type="button"
                                    onClick={() => {
                                        void getStaff();
                                        setIsAddClassModalOpen(!isAddClassModalOpen);
                                    }}
                                >
                                    Добавить класс
                                </button>
                            </div>
                        )}

                        {selectedList === "events" && role === "Owner" && (<button
                            className="btn btn--primary"
                            type="button"
                            onClick={() => {
                                void getEvents
                                setIsAddEventModalOpen(true);
                            }}
                        >
                            Добавить мероприятие
                        </button>)}

                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => setSelectedList("classes")}
                            disabled={selectedList === "classes"}
                        >
                            Классы
                        </button>

                        {role === "Owner" && (
                            <button
                                className="btn btn--secondary"
                                type="button"
                                onClick={() => {
                                    void getStaff();
                                    setSelectedList("staff");
                                }}
                                disabled={selectedList === "staff"}
                            >
                                Персонал
                            </button>
                        )}

                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => {
                                void getEvents();
                                setSelectedList("events");
                            }}
                            disabled={selectedList === "events"}
                        >
                            Мероприятия
                        </button>

                        <button
                            className="btn btn--primary"
                            type="button"
                            onClick={() => navigate("/profile")}
                            disabled={isLoading}
                        >
                            Профиль
                        </button>
                    </div>
                </div>

                <div style={{ height: 16 }} />

                {isLoading && (
                    <div className="class-list">
                        <div className="skeleton" style={{ height: 88 }} />
                        <div className="skeleton" style={{ height: 88 }} />
                        <div className="skeleton" style={{ height: 88 }} />
                    </div>
                )}

                {error && !isLoading && (
                    <div className="alert alert--danger mb-4">{error}</div>
                )}

                {!isLoading && !error && selectedList === "classes" && classes.length > 0 && (
                    <div className="class-list">
                        {classes.map((item) => (
                            <ClassCard key={item.Id} item={item} onClick={() => {
                                navigate("/classDashboard", {
                                    state: {
                                        name: item.Name,
                                        id: item.Id,
                                    },
                                });
                            }}/>
                        ))}
                    </div>
                )}

                {!isLoading && !error && selectedList === "classes" && classes.length === 0 && (
                    <div className="empty-state">
                        <h2 className="empty-state__title">Классы не найдены</h2>
                        <p className="empty-state__text">
                            В системе пока нет доступных классов
                        </p>
                    </div>
                )}

                {!isLoading && !error && selectedList === "events" && classes.length > 0 && (
                    <div className="class-list">
                        {events.map((item) => (
                            <EventCard key={item.ID} item={item} />
                        ))}
                    </div>
                )}

                {!isLoading && !error && selectedList === "events" && events.length === 0 && (
                    <div className="empty-state">
                        <h2 className="empty-state__title">Мероприятия не найдены</h2>
                        <p className="empty-state__text">
                            Не удалось найти доступные мероприятия
                        </p>
                    </div>
                )}

                {!isLoading && !error && selectedList === "staff" && classes.length > 0 && (
                    <div className="class-list">
                        {staff.map((item) => (
                            <StaffCard key={item.Id} user={item} />
                        ))}
                    </div>
                )}

                {!isLoading && !error && selectedList === "staff" && staff.length === 0 && (
                    <div className="empty-state">
                        <h2 className="empty-state__title">Персонал не найден</h2>
                        <p className="empty-state__text">
                            Не удалось найти персонал
                        </p>
                    </div>
                )}

                <AddUserModal
                    isOpen={isAddUserModalOpen}
                    onClose={() => setIsAddUserModalOpen(false)}
                    classes={classes}
                    onAddUser={async (dto) => {
                        await addUser(dto);
                        await getClasses();
                        setIsAddUserModalOpen(false);
                    }}
                />
                
                <AddClassModal isOpen={isAddClassModalOpen} onClose={() => setIsAddClassModalOpen(false)} onAddClass={async (dto) => {
                    await addClass(dto);
                    await getClasses();
                    setIsAddClassModalOpen(false);
                }} staff={staff}/>
                
                <AddEventModal isOpen={isAddEventModalOpen} onClose={() => setIsAddEventModalOpen(false)} onEventAdd={async (dto) => {
                 await addEvent(dto);
                 await getEvents();
                 setIsAddEventModalOpen(false);
                }} classes={classes}/>
            </section>
        </main>
    );
}