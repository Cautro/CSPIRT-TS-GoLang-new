import Foundation

struct ClassesModel: Codable {
    let id: Int
    let name: String
    let grade: Int
    let letter: String
    let teacherLogin: String
    let teacher: UserModel
    let firstQuartyComplete: Int
    let secondQuartyComplete: Int
    let thirdQuertyComplete: Int
    let quarterComplete: Int
    let members: [UserModel]
    let userTotalRating: Int
    let classTotalRating: Int
    
    enum CodingKeys: String, CodingKey {
        case id = "Id"
        case name = "Name"
        case grade = "Grade"
        case letter = "Letter"
        case teacherLogin = "TeacherLogin"
        case teacher = "Teacher"
        case firstQuartyComplete = "FirstQuarterComplete"
        case secondQuartyComplete = "SecondQuarterComplete"
        case thirdQuertyComplete = "ThirdQuarterComplete"
        case quarterComplete = "QuarterComplete"
        case members = "Members"
        case userTotalRating = "UserTotalRating" // user1.rating + user2.rating / n(users)
        case classTotalRating = "ClassTotalRating"
    }
}

struct LocalMyClassCache: Codable {
    let myClass: ClassesModel
}

struct ClassesResponse: Codable {
    let classes: [ClassesModel]
    
    enum CodingKeys: String, CodingKey {
        case classes = "Classes"
    }
}


