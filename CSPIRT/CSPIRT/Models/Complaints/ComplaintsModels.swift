import Foundation

struct ComplaintModel: Codable {
    let id: Int
    let targetName: String
    let targetId: Int
    let authorName: String
    let authorId: Int
    let content: String
    let createdAt: String
    
    enum CodingKeys: String, CodingKey {
        case id = "ID"
        case targetName = "TargetName"
        case targetId = "TargetID"
        case authorName = "AuthorName"
        case authorId = "AuthorID"
        case content = "Content"
        case createdAt = "CreatedAt"
    }
}

struct AddComplaintPayload: Encodable {
    let targetId: Int
    let targetName: String
    let content: String
    let createdAt: String
    
    enum CodingKeys: String, CodingKey {
        case targetId = "TargetID"
        case targetName = "TargetName"
        case content = "Content"
        case createdAt = "CreatedAt"
    }
}

struct LocalComplaintsCache: Codable {
    let complaints: [ComplaintModel]
}

struct EmptyResponse: Decodable {}

