import type {ScheduleLessonType} from "../../entities/schedule/types/schedule_types";
import {useState} from "react";
import {
    ChangeScheduleLessonModal
} from "../../../features/schedule/ui/components/change_schedule_modal.tsx";
import {useAuthStore} from "../../../features/auth/store/auth_store.ts";

interface Props {
    item: ScheduleLessonType;
    lessonNumber: number;
    onChangeScheduleLesson?: () => void;
    type: "base" | "current" | "planned";
    isTeacher?: boolean;
}

export function ScheduleLessonCard({item, lessonNumber, onChangeScheduleLesson = () => console.log(1), type, isTeacher = false}: Props) {
    const [isChangeModalOpen, setIsChangeModalOpen] = useState(false);
    const teacherName = item.Teacher ? `${item.Teacher.Name} ${item.Teacher.LastName}` : "Не указан";
    const role = useAuthStore((state) => state.user?.User.Role);
    const normalizedRole = role?.toLowerCase();

    return (
        <article className="schedule-lesson-card" onClick={() => {
            if (normalizedRole === "owner") setIsChangeModalOpen(true)
        }}>
            <div className="schedule-lesson-card__number">
                {lessonNumber}
            </div>

            <div className="schedule-lesson-card__content">
                <div className="schedule-lesson-card__main">
                    <h3 className="schedule-lesson-card__subject">
                        {item.Subject}
                    </h3>

                    <div className="schedule-lesson-card__time">
                        {item.StartTime} — {item.EndTime}
                    </div>
                </div>

                <div className="schedule-lesson-card__meta">
                    <div className="schedule-lesson-card__metric">
                        <span className="schedule-lesson-card__metric-label">
                            {!isTeacher ? "Учитель" : "Класс"}
                        </span>

                        <span className="schedule-lesson-card__metric-value">
                            {!isTeacher ? teacherName : item.Class}
                        </span>
                    </div>

                    <div className="schedule-lesson-card__metric">
                        <span className="schedule-lesson-card__metric-label">
                            Кабинет
                        </span>
                        <span className="schedule-lesson-card__metric-value">
                            {item.Room || "Не указан"}
                        </span>
                    </div>
                </div>
            </div>

            {!isTeacher && (<ChangeScheduleLessonModal
                lesson={item}
                onChanged={() => {
                    setIsChangeModalOpen(false);
                    onChangeScheduleLesson();
                }}
                isOpen={isChangeModalOpen}
                onClose={() => setIsChangeModalOpen(false)}
                type={type}
            />)}
        </article>
    );
}