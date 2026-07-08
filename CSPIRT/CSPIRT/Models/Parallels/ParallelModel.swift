import Foundation

struct ParallelModel: Codable {
    let id: Int
    let name: String
    let bestClassId: Int?
    let classesIds: [Int]
    let classTotalRating: Int
    
    enum CodingKeys: String, CodingKey {
        case id = "Id"
        case name = "Name"
        case bestClassId = "BestClassId"
        case classesIds = "ClassesIds"
        case classTotalRating = "ClassTotalRating"
    }
}

struct ParallelResponse: Codable {
    let parallels: [ParallelModel]
    
    enum CodingKeys: String, CodingKey {
        case parallels = "ParallelClasses"
    }
}

struct LocalParallelCache: Codable {
    let parallels: [ParallelModel]
    let bestClass: ClassesModel?
    let classes: [ClassesModel]
}

struct SingleParallelResponse: Codable {
    let parallelClass: ParallelModel
    
    enum CodingKeys: String, CodingKey {
        case parallelClass = "ParallelClass"
    }
}
