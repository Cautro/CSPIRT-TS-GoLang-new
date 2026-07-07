import {NoteCard} from "../../../../shared/ui/cards/note_card.tsx";
import {useAuthStore} from "../../../auth/store/auth_store.ts";
import {useNotes} from "../../hooks/use_notes.ts";
import {useDeleteNote} from "../../hooks/use_delete_note.ts";

type props = { id: number, name: string, }

export function NotesWidget({id, name}: props) {
    const getNotes = useNotes(id);
    const notes = getNotes.data;
    const deleteNote = useDeleteNote();
    const role = useAuthStore((state) => state.user?.User.Role);
    
    const isLoading = getNotes.isLoading;
    const error = getNotes.error?.message || deleteNote.error?.message || null;
    
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
            
            {(notes && notes.length > 0) ? (
                <div className={"class-list"}>
                    {notes?.map((note) => (
                        <NoteCard item={note} key={note.ID} onDelete={async () => {
                            await deleteNote.mutateAsync({id: note.ID});
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