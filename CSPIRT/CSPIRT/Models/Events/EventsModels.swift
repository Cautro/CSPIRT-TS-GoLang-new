// EventModels.swift
import Foundation

// MARK: - Существующие модели (дополненные для работы API)

struct EventModel: Codable, Identifiable, Hashable {
    let id: Int
    var title: String
    var status: String
    var baseRatingReward: Int
    var description: String
    var createdAt: String?
    var startedAt: String
    var players: [Int]
    var classes: [Int]
    
    var isActive: Bool { status == "active" }
    
    var statusDisplayName: String {
        switch status {
        case "scheduled", "pending": return "Ожидается"
        case "active": return "Активно"
        case "completed": return "Завершено"
        default: return status.capitalized
        }
    }
    
    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case title = "Title"
        case status = "Status"
        case baseRatingReward = "RatingReward"
        case description = "Description"
        case createdAt = "CreatedAt"
        case startedAt = "StartedAt"
        case players = "Players"
        case classes = "Classes"
    }
}

struct SafeUserModel: Codable, Identifiable, Hashable {
    let id: Int
    let name: String
    let fullName: [FullName] 
    let lastName: String
    let login: String
    let rating: Int
    let role: String
    let classId: Int?
    
    enum CodingKeys: String, CodingKey {
        case id = "Id"
        case name = "Name"
        case fullName = "FullName"
        case lastName = "LastName"
        case login = "Login"
        case rating = "Rating"
        case role = "Role"
        case classId = "ClassID"
    }

    static func == (lhs: SafeUserModel, rhs: SafeUserModel) -> Bool {
        return lhs.id == rhs.id &&
               lhs.name == rhs.name &&
               lhs.lastName == rhs.lastName &&
               lhs.login == rhs.login &&
               lhs.rating == rhs.rating &&
               lhs.role == rhs.role &&
               lhs.classId == rhs.classId
    }

    func hash(into hasher: inout Hasher) {
        hasher.combine(id)
        hasher.combine(name)
        hasher.combine(lastName)
        hasher.combine(login)
        hasher.combine(rating)
        hasher.combine(role)
        hasher.combine(classId)
    }
}


struct UserData: Codable {
    let id: Int
    let login: String
    let role: String
    let classId: Int
    
    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case login = "Login"
        case role = "Role"
        case classId = "ClassID"
    }
}

// MARK: - DTO для запросов управления мероприятиями

struct EventPayload: Codable {
    let title: String
    let status: String
    let ratingReward: Int
    let description: String
    let startedAt: String
    let players: [Int]
    let classes: [Int]
    
    enum CodingKeys: String, CodingKey {
        case title = "Title"
        case status = "Status"
        case ratingReward = "RatingReward"
        case description = "Description"
        case startedAt = "StartedAt"
        case players = "Players"
        case classes = "Classes"
    }
}

struct EventCompletePayload: Codable {
    let ratingReward: Int
    let classReward: Int
}

struct EventPlayersPayload: Codable {
    let playerIds: [Int]
}
