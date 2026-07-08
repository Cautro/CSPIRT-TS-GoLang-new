import Foundation
import Combine

struct LocalDashboardCache: Codable {
    let userName: String
    let me: UserModel
    let topClassUsers: [UserModel]
    let availableEvents: [EventModel]
    let latestNotes: [NoteModel]
    let latestComplaints: [ComplaintModel]
}
