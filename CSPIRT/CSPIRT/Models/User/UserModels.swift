import Foundation

enum UserRole: String, Codable {
    case user = "User"
    case helper = "Helper"
    case admin = "Admin"
    case owner = "Owner"
    case publicRole = "Public"
    case systemAdmin = "SystemAdmin"
}

extension UserRole {
    var displayName: String {
        switch self {
        case .user: return "Ученик"
        case .helper: return "Староста"
        case .admin: return "Учитель"
        case .owner: return "Руководство"
        case .publicRole: return "Публичный"
        case .systemAdmin: return "Системный администратор"
        }
    }
}

struct FullName: Codable, Hashable {
    let Name: String
    let LastName: String
    let MiddleName: String
    
    enum CodingKeys: String, CodingKey {
        case Name = "Name"
        case LastName = "LastName"
        case MiddleName = "MiddleName"
    }
}

struct UserModel: Codable, Identifiable {
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
    
    enum CodingKeys: String, CodingKey {
        case id = "Id"
        case avatar = "Avatar"
        case name = "Name"
        case lastName = "LastName"
        case fullName = "FullName"
        case login = "Login"
        case rating = "Rating"
        case role = "Role"
        case className = "Class"
        case classId = "ClassID"
    }
    
    var primaryFullName: FullName? {
        return fullName.first
    }
}

struct SQLNullString: Codable {
    let string: String
    let valid: Bool
    
    enum CodingKeys: String, CodingKey {
        case string = "String"
        case valid = "Valid"
    }
    
    var value: String? {
        return valid ? string : nil
    }
}

struct ClassUsersResponse: Decodable {
    let users: [UserModel]
    
    enum CodingKeys: String, CodingKey {
        case users = "Users"
    }

}
