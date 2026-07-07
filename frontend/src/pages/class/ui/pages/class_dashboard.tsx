import {useState} from "react";
import {useParams, useSearchParams} from "react-router-dom";
import {useAuthStore} from "../../../../features/auth/store/auth_store.ts";
import {ChangeTeacherModal} from "../components/change_teacher_modal.tsx";
import {ConfirmModal} from "../../../../shared/ui/modals/confirm_modal.tsx";
import {UsersWidget} from "../../../../features/users/ui/components/users_widget.tsx";
import {NotesWidget} from "../../../../features/notes/ui/components/notes_widget.tsx";
import {ComplaintsWidget} from "../../../../features/complaints/ui/components/complaints_widget.tsx";
import {ScheduleWidget} from "../../../../features/schedule/ui/components/schedule_widget.tsx";
import {BaseScheduleWidget} from "../../../../features/schedule/ui/components/base_schedule_widget.tsx";
import {type BurgerDrawerMenuItem} from "../../../../shared/ui/other/burger_menu.tsx";
import {PageHeader} from "../../../../shared/ui/other/page_header.tsx";
import {TabsSwitcher, type TabsSwitcherItem} from "../../../../shared/ui/other/tabs_switcher.tsx";
import {useStaff} from "../../../../features/users/hooks/use_staff.ts";
import {useChangeTeacher} from "../../../../features/class/hooks/use_change_teacher.ts";
import {useClassTeacher} from "../../../../features/class/hooks/use_class_teacher.ts";
import {useDeleteClass} from "../../../../features/class/hooks/use_delete_class.ts";
import {UseRolloverSchedule} from "../../../../features/schedule/hooks/use_rollover_schedule.ts";
import {AddUserModal} from "../../../../features/users/ui/components/add_user_modal.tsx";
// import {PlannedScheduleWidget} from "../../../../features/schedule/ui/components/planned_schedule_widget.tsx";

type SelectedList = | "users" | "notes" | "complaints" | "schedule" | "baseschedule" | "plannedschedule";

export function ClassDashboard() {
    const [searchParams, setSearchParams] = useSearchParams();
    
    const { id } = useParams<{id: string, name: string}>();
    const name = searchParams.get("name");
    const classId = id ? parseInt(id, 10) : null;
    const role = useAuthStore((state) => state.user?.User.Role);
    const normalizedRole = role?.toLowerCase();
    const changeTeacher = useChangeTeacher();
    const getStaff = useStaff();
    const getClassTeacher = useClassTeacher(classId ?? 0);
    const deleteClass = useDeleteClass()
    const teacher = getClassTeacher.data
    const rolloverSchedule   = UseRolloverSchedule();
    const error = changeTeacher.error?.message || getStaff.error?.message || deleteClass.error?.message || getStaff.error?.message || getClassTeacher.error?.message;
    
    
    const isLoading = changeTeacher.isPending || getStaff.isLoading || deleteClass.isPending || getStaff.isLoading || getClassTeacher.isLoading;


    const selectedList =
        (searchParams.get("tab") as SelectedList) || "users";

    const setSelectedList = (tab: SelectedList) => {
        setSearchParams(prev => {
            prev.set("tab", tab);
            return prev;
        }, { replace: true });
    };
    const [isChangeTeacherModalOpen, setChangeTeacherModalOpen] = useState(false);
    const [isDeleteClassModalOpen, setDeleteClassModalOpen] = useState(false);
    const [isRolloverModalOpen, setRolloverModalOpen] = useState(false);
    const [isAddUserModalOpen, setIsAddUserModalOpen] = useState(false);
    const [key, setKey] = useState<number>(0);

    const menuItems: BurgerDrawerMenuItem[] = [
        {
            label: "Добавить пользователя",
            onClick: () => setIsAddUserModalOpen(true),
            hidden: (normalizedRole !== "owner" && normalizedRole !== "admin"),
        },
        {
            label: "Изменить классного руководителя",
            onClick: async () => {
                await getStaff.refetch();
                setChangeTeacherModalOpen(true);
            },
            hidden: normalizedRole !== "owner",
        },
        {
            label: "Удалить класс",
            onClick: () => setDeleteClassModalOpen(true),
            hidden: normalizedRole !== "owner",
        },
        {
            label: "Сбросить расписание",
            onClick: () => setRolloverModalOpen(true),
            hidden: normalizedRole !== "owner",
        },
    ];
    
    const tabs: TabsSwitcherItem<SelectedList>[] = [
        {
            value: "users",
            label: "Ученики",
        },
        {
            value: "schedule",
            label: "Расписание",
        },
        {
            value: "baseschedule",
            label: "Стандартное расписание",
            hidden: normalizedRole !== "owner",
        },
        {
            value: "notes",
            label: "Заметки",
            hidden: !(normalizedRole === "admin" || normalizedRole === "owner" || normalizedRole === "helper"),
        },
        {
            value: "complaints",
            label: "Жалобы",
            hidden: !(normalizedRole === "admin" || normalizedRole === "owner"),
        },
    ]

    return (
        <main className={"main"}>
            <section className={"page"}>
                
                <PageHeader 
                    title={`${name} Класс`}
                    description={`Классный руководитель - ${teacher?.Name ?? ""} ${teacher?.LastName ?? ""}`}
                    menuItems={menuItems}
                    hasBackButton={true}
                />
                
                <div className="page-spacer"></div>
                
                <TabsSwitcher
                    items={tabs}
                    value={selectedList}
                    onChange={setSelectedList}
                />

                <div className="page-spacer"></div>
                
                {isLoading && (
                    <div className="grid grid--3">
                        <div className="skeleton" style={{ height: 160 }} />
                        <div className="skeleton" style={{ height: 160 }} />
                        <div className="skeleton" style={{ height: 160 }} />
                    </div>
                )}

                {error && !isLoading && (
                    <div className="alert alert--danger mb-4">{error}</div>
                )}
                
                {selectedList === "users" && (
                    <UsersWidget name={name ?? ""} id={classId ?? 0} key={key} />
                )}
                
                {selectedList === "notes" && (
                    <NotesWidget id={classId ?? 0} name={name ?? ""} key={key} />
                )}

                {selectedList === "complaints" && (
                    <ComplaintsWidget id={classId ?? 0} name={name ?? ""} key={key} />
                )}

                {selectedList === "schedule" && (
                    <ScheduleWidget id={classId ?? 0} name={name ?? ""} key={key} />
                )}

                {selectedList === "baseschedule" && (
                    <BaseScheduleWidget id={classId ?? 0} name={name ?? ""} key={key} />
                )}

                {/*{selectedList === "plannedschedule" && (*/}
                {/*    <PlannedScheduleWidget id={classId ?? 0} name={name ?? ""} key={key} />*/}
                {/*)}*/}
                
            </section>
            
            <ChangeTeacherModal isOpen={isChangeTeacherModalOpen} onClose={() => setChangeTeacherModalOpen(false)} onChangeTeacher={async (dto) => {
                if (classId !== null) {
                    await changeTeacher.mutateAsync({id: classId, teacher: dto});
                    getClassTeacher.refetch();
                    setKey(key + 1);
                    setChangeTeacherModalOpen(false);
                } 
            }}className={name ?? ""}/>
            
            <ConfirmModal
                title={"Удалить класс?"}
                content={`Это действие удалит ${name} класс. Отменить удаление будет нельзя.`}
                onConfirm={async () => {
                    if (classId !== null) {
                        await deleteClass.mutateAsync({id: classId});
                        setKey(key + 1);
                        setDeleteClassModalOpen(false);
                    }
                }}
                isOpen={isDeleteClassModalOpen}
                onClose={() => setDeleteClassModalOpen(false)}
                buttonContent={"Удалить"}
                isDanger={true}
            />

            <ConfirmModal
                title={"Сброс расписания"}
                content={"Вы уверены что хотите сбросить текущее расписание на стандартное?"}
                buttonContent={"Сбросить"}
                onConfirm={async () => {
                    if (classId) {
                        await rolloverSchedule.mutateAsync({id: classId ?? 0});
                        setKey(key + 1);
                        setRolloverModalOpen(false);
                    }
                }}
                isOpen={isRolloverModalOpen}
                onClose={() => setRolloverModalOpen(false)}
            />

            <AddUserModal
                isOpen={isAddUserModalOpen}
                onClose={() => setIsAddUserModalOpen(false)}
                onAddUser={async () => {
                    setKey(prevKey => prevKey + 1);
                    setIsAddUserModalOpen(false);
                }}
                classId={classId}
            />
            
        </main>
    );
}
