import Foundation

struct NoteModel: Codable {
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

struct LocalNoteCache: Codable {
    let notes: [NoteModel]
}

struct AddNotePayload: Codable {
    let targetId: Int
    let targetName: String
    let content: String
    let createdAt: String
}
