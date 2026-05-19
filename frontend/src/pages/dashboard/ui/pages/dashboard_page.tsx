import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "../../../../features/auth/store/auth_store.ts";
import { AddUserModal } from "../../../../features/users/ui/components/add_user_modal.tsx";
import {AddClassModal} from "../../../../features/class/ui/components/add_class_modal.tsx";
import {AddEventModal} from "../../../../features/events/ui/components/add_event_modal.tsx";
import {ClassesWidget} from "../../../../features/class/ui/components/classes_widget.tsx";
import {EventsWidget} from "../../../../features/events/ui/components/events_widget.tsx";
import {StaffWidget} from "../../../../features/users/ui/components/staff_widget.tsx";
import {BurgerDrawerMenu, type BurgerDrawerMenuItem} from "../../../../shared/ui/other/burger_menu.tsx";

type Lists = "classes" | "events" | "staff";

export function DashboardPage() {
    const navigate = useNavigate();

    const role = useAuthStore((state) => state.user?.User.Role);
    
    const [selectedList, setSelectedList] = useState<Lists>("classes");
    const [isAddUserModalOpen, setIsAddUserModalOpen] = useState(false);
    const [isAddClassModalOpen, setIsAddClassModalOpen] = useState(false);
    const [isAddEventModalOpen, setIsAddEventModalOpen] = useState(false);
    const [key, setKey] = useState(0);
    
    const menuItems: BurgerDrawerMenuItem[] = [
        {
            label: "Список классов",
            onClick: () => setSelectedList("classes"),
            disabled: selectedList === "classes"
        },
        {
            label: "Персонал школы",
            onClick: () => setSelectedList("staff"),
            disabled: selectedList === "staff",
            hidden: (role !== "Owner")
        },
        {
            label: "Мероприятия",
            onClick: () => setSelectedList("events"),
            disabled: selectedList === "events",
        },
        {
            label: "Добавить пользователя",
            primary: true,
            onClick: () => setIsAddUserModalOpen(true),
            hidden: (role !== "Owner" || selectedList !== "classes"),
        },
        {
            label: "Добавить класс",
            primary: true,
            onClick: () => setIsAddClassModalOpen(true),
            hidden: (role !== "Owner" || selectedList !== "classes"),
        },
        {
            label: "Добавить мероприятие",
            primary: true,
            onClick: () => setIsAddEventModalOpen(true),
            hidden: (role !== "Owner" || selectedList !== "events"),
        },
        {
            label: "Профиль",
            primary: true,
            onClick: () => navigate("/profile")
        },
    ]

    if (!role) {
        return null;
    }

    return (
        <main className="main">
            <section className="page">
                <div className="profile-hero class-dashboard-hero">
                    <div className="class-dashboard-hero__content">
                        <h1 className="info-row__value">Панель управления</h1>
                        
                    </div>

                    <div className="class-dashboard-hero__menu">
                        <BurgerDrawerMenu
                            title="Меню"
                            items={menuItems}
                            side="right"
                        />
                    </div>
                </div>

                <div className="page-spacer" />

                {selectedList === "classes" && (
                    <ClassesWidget key={key}/>
                )}

                {selectedList === "events" && (
                    <EventsWidget key={key}/>
                )}

                {selectedList === "staff" && (
                    <StaffWidget key={key}/>
                )}

                <AddUserModal
                    isOpen={isAddUserModalOpen}
                    onClose={() => setIsAddUserModalOpen(false)}
                    onAddUser={async () => {
                        setKey(prevKey => prevKey + 1);
                        setIsAddUserModalOpen(false);
                    }}
                />
                
                <AddClassModal isOpen={isAddClassModalOpen} onClose={() => setIsAddClassModalOpen(false)} onAddClass={async () => {
                    setKey(prevKey => prevKey + 1);
                    setIsAddClassModalOpen(false);
                }}/>
                
                <AddEventModal isOpen={isAddEventModalOpen} onClose={() => setIsAddEventModalOpen(false)} onEventAdd={async () => {
                 setKey(prevKey => prevKey + 1);
                 setIsAddEventModalOpen(false);
                }} />
            </section>
        </main>
    );
}
