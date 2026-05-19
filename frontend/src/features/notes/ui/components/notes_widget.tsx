import {useEffect} from "react";
import {useNotesStore} from "../../store/notes_store.ts";
import {NoteCard} from "../../../../shared/ui/cards/note_card.tsx";
import {useAuthStore} from "../../../auth/store/auth_store.ts";

type props = {
    id: number,
    name: string,
}

export function NotesWidget({id, name}: props) {
    const notes = useNotesStore((state) => state.notes)
    const getNotes = useNotesStore((state) => state.getNotes)
    const deleteNote = useNotesStore((state) => state.deleteNote);
    const role = useAuthStore((state) => state.user?.User.Role);
    const status = useNotesStore((state) => state.status);
    const error = useNotesStore((state) => state.error);
    
    const isLoading = status === "loading";

    useEffect(() => {
        void getNotes(id);
    }, [getNotes, id])

    return (
        <>
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
            
            {(notes !== null && notes.length > 0) ? (
                <div className={"class-list"}>
                    {notes.map((note) => (
                        <NoteCard item={note} key={note.ID} onDelete={async () => {
                            await deleteNote(note.ID);
                            getNotes(id);
                        }} role={role ?? "User"} />
                    ))}
                </div>
            ) :(
                !isLoading && <div className="empty-state">
                    <h2 className="empty-state__title">Заметки не найдены</h2>
                    <p className="empty-state__text">
                        Не удалось найти заметки по {name} классу
                    </p>
                </div>
            )}
        </>
    );
} 