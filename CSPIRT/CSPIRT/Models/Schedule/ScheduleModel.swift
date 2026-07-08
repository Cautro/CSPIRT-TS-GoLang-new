import Combine
import Foundation

struct ScheduleModel: Codable, Identifiable {
    let id: Int
    let type: String
    let baseScheduleId: Int?
    let classId: Int
    let className: String
    let dayOfWeek: String
    let lessonNumber: Int
    let weekType: String
    let subject: String
    let teacherId: Int
    let teacher: UserModel
    let room: Int
    let startTime: String
    let endTime: String
    let description: String
    let createdAt: String?
    
    enum CodingKeys: String, CodingKey {
        case id = "Id"
        case type = "Type"
        case baseScheduleId = "BaseScheduleID"
        case classId = "ClassID"
        case className = "Class"
        case dayOfWeek = "DayOfWeek"
        case lessonNumber = "LessonNumber"
        case weekType = "WeekType"
        case subject = "Subject"
        case teacherId = "TeacherID"
        case teacher = "Teacher"
        case room = "Room"
        case startTime = "StartTime"
        case endTime = "EndTime"
        case description = "Description"
        case createdAt = "CreatedAt"
    }
}

struct SchedulesResponse: Codable {
    let schedules: [ScheduleModel]
    let base: [ScheduleModel]
    let current: [ScheduleModel]
    let planned: [ScheduleModel]
    
    enum CodingKeys: String, CodingKey {
        case schedules = "Schedules"
        case base = "Base"
        case current = "Current"
        case planned = "Planned"
    }
}
