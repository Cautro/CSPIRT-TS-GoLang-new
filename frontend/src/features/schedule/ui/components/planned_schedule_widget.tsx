import {ScheduleDayCard} from "../../../../shared/ui/cards/schedule_day_card.tsx";
import {useClassSchedule} from "../../hooks/use_class_schedule.ts";
import {useEffect} from "react";

type props = {
    id: number,
    name: string,
}

export function PlannedScheduleWidget({id, name}: props) {
    const {data: schedule, error: error, isLoading: isLoading, refetch: refetch} = useClassSchedule(id, "planned");

    useEffect(() => {
        refetch();
    }, [id])
    
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
                <div className="alert alert--danger mb-4">{error.message}</div>
            )}

            {schedule ? (
                <div className="schedule-days-list">
                    <ScheduleDayCard title="Понедельник" lessons={schedule.monday ?? []} onChangeScheduleLesson={async () => refetch()} classId={id} day={"monday"}  type={"planned"} />
                    <ScheduleDayCard title="Вторник" lessons={schedule.tuesday ?? []} onChangeScheduleLesson={async () => refetch()}  classId={id} day={"tuesday"} type={"planned"} />
                    <ScheduleDayCard title="Среда" lessons={schedule.wednesday ?? []} onChangeScheduleLesson={async () => refetch()} classId={id} day={"wednesday"} type={"planned"} />
                    <ScheduleDayCard title="Четверг" lessons={schedule.thursday ?? []} onChangeScheduleLesson={async () => refetch()} classId={id} day={"thursday"} type={"planned"} />
                    <ScheduleDayCard title="Пятница" lessons={schedule.friday ?? []} onChangeScheduleLesson={async () => refetch()} classId={id} day={"friday"} type={"planned"} />
                </div>
            ) : !isLoading && !schedule && (
                <div className="empty-state">
                    <h2 className="empty-state__title">Расписание не найдено</h2>
                    <p className="empty-state__text">
                        Не удалось найти расписание для {name} класса
                    </p>
                </div>
            )}

        </>
    );
} 