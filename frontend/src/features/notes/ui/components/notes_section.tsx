import {NoteCard} from "../../../../shared/ui/cards/note_card.tsx";
import type {NoteType} from "../../../../shared/entities/notes/types/notes_types.ts";
import {useState} from "react";
import {useAddNote} from "../../hooks/use_add_note.ts";
import type {UserType} from "../../../../shared/entities/user/types/user_types.ts";
import {useDeleteNote} from "../../hooks/use_delete_note.ts";
import {useUser} from "../../../users/hooks/use_user.ts";
import {NoteAddUsecase} from "../../models/note_add_usecase.ts";

export const NotesSection = ({isYou, notes, setFormError = () => {}, currentUser, user, isProfile = false}: {isProfile?: boolean, user: UserType, currentUser: UserType, isYou: boolean, notes: NoteType[], setFormError?: (value: string | null) => void}) => {
  const [noteText, setNoteText] = useState("");
  const addNote = useAddNote()
  const deleteNote = useDeleteNote()
  const getUser = useUser(user.Id)
  const [isSubmiting, setIsSubmitting] = useState(false);

  async function handleNoteAdd() {
    setFormError(null);
    setIsSubmitting(true);

    if (!user || !currentUser) {
      setFormError("Не удалось определить пользователя");
      return;
    }

    const form = {
      user: user,
      current_user: currentUser,
      content: noteText,
    };
      
    
    try {
      const dto = NoteAddUsecase(form);
      await addNote.mutateAsync({form: dto})
      setNoteText("");
      getUser.refetch()
    } catch (e) {
      setFormError(e instanceof Error ? e.message : "Неизвестная ошибка");
    }
    setIsSubmitting(false);
  }
  
  return (
      <section className="card card--padded user-section-card">
        <div className="section-head section-head--row">
          <div>
            <h2 className="section-title">Заметки</h2>
            <p className="section-description">
              Поведенческие заметки, оставленные ответственными пользователями.
            </p>
          </div>

          <span className="badge badge--neutral">{notes.length}</span>
        </div>

        {(isProfile === false && !isYou) && (<div className="user-action-form">
          <div className="field">
            <label className="field__label" htmlFor="noteText">
              Новая заметка
            </label>

            <textarea
                id="noteText"
                className="textarea"
                placeholder="Введите заметку о поведении пользователя..."
                maxLength={500}
                value={noteText}
                onChange={(event) => setNoteText(event.target.value)}
            />
          </div>

          <div className="user-action-form__footer">
            <button
                className="btn btn--primary"
                type="button"
                disabled={!noteText.trim() || isProfile}
                onClick={() => void handleNoteAdd()}
            >
              {!isSubmiting ? "Добавить заметку" : "Заметка добавляется"}
            </button>
          </div>
        </div>)}

        {isProfile === false ? (notes.length > 0 ? (
            <div className="feed">
              {notes.map((note) => (
                  <NoteCard
                      key={note.ID}
                      item={note}
                      onDelete={async () => {
                        await deleteNote.mutateAsync({id: note.ID});
                        getUser.refetch();
                        console.log(getUser.data);
                      }}
                      role={currentUser.Role}
                  />
              ))}
            </div>
        ) : (
            <div className="empty-inline">Заметок нет</div>
        )) : (notes.length > 0 ? (
            <div className="feed">
              {notes.map((note) => (
                  <NoteCard
                      key={note.ID}
                      item={note}
                      role={currentUser.Role}
                  />
              ))}
            </div>
        ) : (
            <div className="empty-inline">Заметок нет</div>
        ))}
      </section>
  )
}