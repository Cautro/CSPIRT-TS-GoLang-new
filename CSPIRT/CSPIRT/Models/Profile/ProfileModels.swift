import Foundation

struct UserProfileModel: Decodable {
    let id: Int?
    let avatar: SQLNullString?
    let name: String
    let lastName: String
    let fullName: [FullName]
    let login: String
    let rating: Int
    let role: UserRole
    let className: String
    let classId: Int
}

struct MeResponse: Codable {
    let user: UserModel
    let notes: [NoteModel]
    let complaints: [ComplaintModel]
    let classTeacher: UserModel?
    let events: [EventModel]
    
    enum CodingKeys: String, CodingKey {
        case user = "User"
        case notes = "Notes"
        case complaints = "Complaints"
        case classTeacher = "ClassTeacher"
        case events = "Events"
    }
}

struct UpdateAvatarResponse: Decodable {
    let Avatar: String?
    
    enum CodingKeys: String, CodingKey {
        case Avatar = "avatar"
    }
}

struct LocalProfileModel: Codable {
    let user: UserModel
//    let notes: [NoteModel]
//    let complaints: [ComplaintModel]
//    let classTeacher: UserModel?
//    let events: [EventModel]
}

struct UpdateRatingResponse: Codable {
    let message: String
    let target: String
    let newRating: Int
    
    enum CodingKeys: String, CodingKey {
        case message = "message"
        case target = "target"
        case newRating = "new_rating"
    }
}
