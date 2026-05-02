import {useEffect, useState} from "react";
import {useLocation, useNavigate} from "react-router-dom";
import {UserCard} from "../../../../shared/ui/user_card.tsx";
import {NoteCard} from "../../../../shared/ui/note_card.tsx";
import {useClassDashboardStore} from "../../store/class_dashboard_store.ts";
import {ComplaintCard} from "../../../../shared/ui/complaint_card.tsx";
import {useAuthStore} from "../../../auth/store/auth_store.ts";

type SelectedList = | "users" | "notes" | "complaints";

export function ClassDashboard() {
    const navigate = useNavigate();
    
    const location = useLocation();
    const className = location.state?.name;
    const classId = location.state?.id;
    const role = useAuthStore((state) => state.user?.User.Role);
    
    const getUsers = useClassDashboardStore((state) => state.getUsersByClass);
    const users = useClassDashboardStore((state) => state.users);
    const status = useClassDashboardStore((state) => state.status);
    const error = useClassDashboardStore((state) => state.error);
    const notes = useClassDashboardStore((state) => state.notes);
    const getNotes = useClassDashboardStore((state) => state.getNotesByClass);
    const deleteNote = useClassDashboardStore((state) => state.deleteNote);
    const complaints = useClassDashboardStore((state) => state.complaints);
    const getComplaints = useClassDashboardStore((state) => state.getComplaints);
    const deleteComplaint = useClassDashboardStore((state) => state.deleteComplaint);
    
    const isLoading = status === "loading";
    
    const [selectedList, setSelectedList] = useState<SelectedList>("users"); 
        
    useEffect(() => {
        void getUsers(classId);
    }, [getUsers, classId])

    return (
        <main className={"main"}>
            <section className={"page"}>
                <div className={"page__head"}>
                    <div>
                        <h1 className={"page__title"}>{className} Класс</h1>
                        <p className={"page__description"}>Посмотрите список учеников конкретного класса и информацию о них</p>
                    </div>
                    
                    <div className={"btn-group"}>
                        <button
                            className={"btn btn--secondary"}
                            type={"button"}
                            onClick={() => setSelectedList('users')}
                            disabled={(selectedList === "users")}
                        >
                            Список учеников
                        </button>

                        {role === "Admin" || role === "Owner" || role === "Helper" ? (
                            <button
                                className={"btn btn--secondary"}
                                type={"button"}
                                onClick={() => {
                                    void getNotes(classId);
                                    setSelectedList('notes');
                                }}
                                disabled={(selectedList === "notes")}
                            >
                                Список заметок класса
                            </button>
                        ) : <div></div>}

                        <button
                            className={"btn btn--secondary"}
                            type={"button"}
                            onClick={() => {
                                setSelectedList('complaints');
                                getComplaints(classId);
                            }}
                            disabled={(selectedList === "complaints")}
                        >
                            Список жалоб класса
                        </button>
                        
                        <button
                            className={"btn btn--primary"}
                            type="button"
                            onClick={() => {
                                navigate("/");
                            }}
                        >
                            На главную
                        </button>
                    </div>
                </div>

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
                
                {selectedList === "users" ? users.length > 0 ? (
                    <div className={"class-list"}>
                        {users.map((user) => (
                            <UserCard user={user} key={user.Id}/>
                        ))}
                    </div>
                ) : (
                    !isLoading && <div className="empty-state">
                        <h2 className="empty-state__title">Ученики не найдены</h2>
                        <p className="empty-state__text">
                            Не удалось найти учеников по {className} классу
                        </p>
                    </div>
                ) : <div></div>}

                    
                
                {selectedList === "notes" && notes.length > 0 ? (
                    <div className={"class-list"}>
                        {notes.map((note) => (
                            <NoteCard item={note} key={note.ID} onDelete={() => {
                                deleteNote(note.ID.toString());
                                if (!classId) {
                                    return
                                }
                                getNotes(classId);
                            }} />
                        ))}
                    </div>
                ) :(
                    selectedList === "notes" && !isLoading && <div className="empty-state">
                        <h2 className="empty-state__title">Заметки не найдены</h2>
                        <p className="empty-state__text">
                            Не удалось найти заметки по {className} классу
                        </p>
                    </div>
                )}

                {selectedList === "complaints" && complaints.length > 0 ? (
                    <div className={"class-list"}>
                        {complaints.map((item) => (
                            <ComplaintCard item={item} key={item.ID} onDelete={() => {
                                deleteComplaint(item.ID.toString());
                                if (!classId) {
                                    return
                                }
                                getComplaints(classId);
                            }} />
                        ))}
                    </div>
                ) :(
                    selectedList === "complaints" && !isLoading && <div className="empty-state">
                        <h2 className="empty-state__title">Жалобы не найдены</h2>
                        <p className="empty-state__text">
                            Не удалось найти жалобы по {className} классу
                        </p>
                    </div>
                )}
            </section>
        </main>
    );
}