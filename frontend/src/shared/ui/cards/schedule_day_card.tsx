import type {ScheduleDay, ScheduleLessonType} from "../../entities/schedule/types/schedule_types";
import { ScheduleLessonCard } from "./schedule_lesson_card";
import {useAuthStore} from "../../../features/auth/store/auth_store.ts";
import {AddScheduleLessonModal} from "../../../features/schedule/ui/components/add_schedule_modal.tsx";
import {useState} from "react";

interface Props {
    title: string;
    lessons: ScheduleLessonType[];
    onChangeScheduleLesson?: () => void;
    classId?: number;
    isTeacher?: boolean;
    day: ScheduleDay;
    type: "base" | "current" | "planned"
}

export function ScheduleDayCard({ title, lessons, onChangeScheduleLesson = () => console.log(1), day, classId, type, isTeacher = false}: Props) {
    const role = useAuthStore((state) => state.user?.User.Role);
    const [isAddLessonModalOpen, setIsAddLessonModalOpen] = useState(false);
    
    return (
        <section className="schedule-day-card">
            <div className="schedule-day-card__header">
                <div>
                    <h2 className="schedule-day-card__title">{title}</h2>
                    <p className="schedule-day-card__subtitle">
                        Расписание уроков на выбранный день
                    </p>
                </div>

                <span className="schedule-day-card__badge">
          {lessons.length} уроков
        </span>
            </div>

            {lessons.length > 0 ? (
                <div className="schedule-day-card__list">
                    {lessons.map((lesson, index) => (
                        <ScheduleLessonCard
                            key={lesson.Id ?? `${lesson.Subject}-${index}`}
                            item={lesson}
                            lessonNumber={index + 1}
                            onChangeScheduleLesson={() => onChangeScheduleLesson()}
                            type={type}
                            isTeacher={isTeacher}
                        />
                    ))}
                </div>
            ) : (
                <div className="empty-inline">
                    На этот день уроков нет
                </div>
            )}
            {role === "Owner" && !isTeacher && (
                <div className="btn-group">
                    <button className="btn btn--primary"
                        onClick={() => setIsAddLessonModalOpen(true)}
                    >
                        Добавить урок
                    </button>
                </div>
            )}

            {!isTeacher && classId && (
                <AddScheduleLessonModal 
                    isOpen={isAddLessonModalOpen}
                    onClose={() => setIsAddLessonModalOpen(false)}
                    classId={classId}
                    dayOfWeek={day}
                    onAdded={() => onChangeScheduleLesson()}
                    type={type}
                />
            )}
            
        </section>
    );
}