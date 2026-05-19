import type {ScheduleLessonType, ScheduleModelType} from "../types/schedule_types.ts";

export const ScheduleService = {
    sortSchedule(data: ScheduleLessonType[]): ScheduleModelType {
        const sorted = [...data].sort((a, b) =>
            a.EndTime.localeCompare(b.EndTime)
        );
        
        return {
            monday: sorted.filter(value => value.DayOfWeek === "monday"),
            tuesday: sorted.filter(value => value.DayOfWeek === "tuesday"),
            wednesday: sorted.filter(value => value.DayOfWeek === "wednesday"),
            thursday: sorted.filter(value => value.DayOfWeek === "thursday"),
            friday: sorted.filter(value => value.DayOfWeek === "friday"),
            saturday: sorted.filter(value => value.DayOfWeek === "saturday"),
            sunday: sorted.filter(value => value.DayOfWeek === "sunday")
        }   
    }
}