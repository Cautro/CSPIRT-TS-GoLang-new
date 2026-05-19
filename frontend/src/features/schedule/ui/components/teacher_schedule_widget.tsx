import {ScheduleDayCard} from "../../../../shared/ui/cards/schedule_day_card.tsx";
import {useScheduleStore} from "../../store/schedule_store.ts";
import {useEffect} from "react";

type props = {
    name: string,
    id: number
}

export function TeacherScheduleWidget({name, id}: props) {
    const schedule = useScheduleStore((state) => state.schedule);
    const getSchedule = useScheduleStore((state) => state.getTeacherSchedule);
    const status = useScheduleStore((state) => state.status);
    const error = useScheduleStore((state) => state.error)

    const isLoading = status === "loading";

    useEffect(() => {
        void getSchedule(id);
    }, [getSchedule, id]);

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

            {schedule ? (
                <div className="schedule-days-list">
                    <ScheduleDayCard title="Понедельник" lessons={schedule.monday ?? []} onChangeScheduleLesson={async () => getSchedule(id)} isTeacher={true} day={"monday"}  type={"current"} />
                    <ScheduleDayCard title="Вторник" lessons={schedule.tuesday ?? []} onChangeScheduleLesson={async () => getSchedule(id)} isTeacher={true} day={"tuesday"} type={"current"} />
                    <ScheduleDayCard title="Среда" lessons={schedule.wednesday ?? []} onChangeScheduleLesson={async () => getSchedule(id)} isTeacher={true} day={"wednesday"} type={"current"} />
                    <ScheduleDayCard title="Четверг" lessons={schedule.thursday ?? []} onChangeScheduleLesson={async () => getSchedule(id)} isTeacher={true} day={"thursday"} type={"current"} />
                    <ScheduleDayCard title="Пятница" lessons={schedule.friday ?? []} onChangeScheduleLesson={async () => getSchedule(id)} isTeacher={true} day={"friday"} type={"current"} />
                </div>
            ) : !isLoading && !schedule && (
                <div className="empty-state">
                    <h2 className="empty-state__title">Расписание не найдено</h2>
                    <p className="empty-state__text">
                        Не удалось найти расписание для {name}
                    </p>
                </div>
            )}

        </>
    );
} 