import { useState } from "react";
import {Navigate, useNavigate, useSearchParams} from "react-router-dom";
import { useAuthStore } from "../../../../features/auth/store/auth_store.ts";
import { AddUserModal } from "../../../../features/users/ui/components/add_user_modal.tsx";
import {AddClassModal} from "../../../../features/class/ui/components/add_class_modal.tsx";
import {AddEventModal} from "../../../../features/events/ui/components/add_event_modal.tsx";
import {ClassesWidget} from "../../../../features/class/ui/components/classes_widget.tsx";
import {EventsWidget} from "../../../../features/events/ui/components/events_widget.tsx";
import {StaffWidget} from "../../../../features/users/ui/components/staff_widget.tsx";
import {type BurgerDrawerMenuItem} from "../../../../shared/ui/other/burger_menu.tsx";
import {PageHeader} from "../../../../shared/ui/other/page_header.tsx";
import {TabsSwitcher, type TabsSwitcherItem} from "../../../../shared/ui/other/tabs_switcher.tsx";
import {UserRound} from "lucide-react";
import {useLogout} from "../../../../features/auth/hooks/use_logout.ts";
import {ParallelsWidget} from "../../../../features/class/ui/components/parallels_widget.tsx";
import {ConfirmModal} from "../../../../shared/ui/modals/confirm_modal.tsx";
import {useCompleteYear} from "../../../../features/class/hooks/use_complete_year.ts";

type Lists = "classes" | "events" | "staff" | "parallels";

export function DashboardPage() {
    const navigate = useNavigate();
    const logout = useLogout();

    const role = useAuthStore((state) => state.user?.User.Role);
    const normalizedRole = role?.toLowerCase();

    const [searchParams, setSearchParams] = useSearchParams();

    const selectedList =
        (searchParams.get("tab") as Lists) || "classes";

    const setSelectedList = (tab: Lists) => {
        setSearchParams({ tab }, {replace: true});
    };
    const [isAddUserModalOpen, setIsAddUserModalOpen] = useState(false);
    const [isAddClassModalOpen, setIsAddClassModalOpen] = useState(false);
    const [isAddEventModalOpen, setIsAddEventModalOpen] = useState(false);
    const [isYearCompleteModalOpen, setIsYearCompleteModalOpen] = useState(false);
    const [key, setKey] = useState(0);
    
    const completeYear = useCompleteYear()
    
    const menuItems: BurgerDrawerMenuItem[] = [
        {
            label: "Добавить пользователя",
            onClick: () => setIsAddUserModalOpen(true),
            hidden: (normalizedRole !== "owner" || selectedList !== "classes"),
        },
        {
            label: "Добавить класс",
            onClick: () => setIsAddClassModalOpen(true),
            hidden: (normalizedRole !== "owner" || selectedList !== "classes"),
        },
        {
            label: "Добавить мероприятие",
            onClick: () => setIsAddEventModalOpen(true),
            hidden: (normalizedRole !== "owner" || selectedList !== "events"),
        },
        {
            label: "Завершить учебный год",
            onClick: () => setIsYearCompleteModalOpen(true),
            hidden: (normalizedRole !== "owner"),
        }
    ]
    
    const tabs: TabsSwitcherItem<Lists>[] = [
        {
            label: "Классы",
            value: "classes",
        },
        {
            label: "Параллели",
            value: "parallels",
        },
        {
            value: "events",
            label: "Мероприятия",
        },
        {
            label: "Персонал",
            value: "staff",
            hidden: normalizedRole !== "owner"
        }
    ]

    if (!role) {
        return null;
    }
    
    if (normalizedRole === "public") {
        return <Navigate to={"/public"}/>
    }

    return (
        <main className="main">
            <section className="page">
                <PageHeader
                    title={"Рейтинг классов МАОУ СОШ 16-Ф"}
                    description={"Просматривайте список классов, мероприятий и прочую информацию"}
                    menuItems={menuItems}
                    menuTitle={"Меню"}
                    actions={
                    <>
                        {normalizedRole !== "public" ? (
                            <button
                                className="app-drawer-button"
                                type="button"
                                onClick={() => navigate("/profile")}
                                aria-label="Перейти в профиль"
                            >
                                <UserRound size={22} />
                            </button>
                        ) : (
                        <button
                            className="btn btn--danger"
                            type="button"
                            onClick={async () => await logout.mutateAsync()}
                            aria-label="Выйти из аккаунта"
                        >
                            Выйти
                        </button>
                        )}
                    </>
                    }
                />

                <div className="page-spacer" />

                <TabsSwitcher
                    items={tabs}
                    value={selectedList}
                    onChange={setSelectedList}
                />
                
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

                {selectedList === "parallels" && (
                    <ParallelsWidget key={key}/>
                )}

                <ConfirmModal 
                    content={"Вы уверены что хотите завершить учебный год?"} 
                    buttonContent={"Завершить"} 
                    isOpen={isYearCompleteModalOpen} 
                    onClose={() => setIsYearCompleteModalOpen(false)} 
                    onConfirm={async () => {
                        await completeYear.mutateAsync()
                        setIsYearCompleteModalOpen(false);
                    }}
                />

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
