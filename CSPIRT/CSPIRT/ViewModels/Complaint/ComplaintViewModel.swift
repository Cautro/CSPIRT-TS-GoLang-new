import Foundation
import Combine

@MainActor
final class ComplaintViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errMsg: String?

    @Published var complaints: [ComplaintModel] = []
    @Published var user: UserModel?

    private let cacheTTL: TimeInterval = 2 * 24 * 60 * 60

    func fetchComplaints(for userId: Int? = nil) async {
        errMsg = nil

        if userId == nil {
            if let cache = await AppCacheStore.shared.load(
                LocalComplaintsCache.self,
                for: .complaints,
                maxAge: cacheTTL
            ) {}
        }

        if complaints.isEmpty {
            isLoading = true
        }

        defer { isLoading = false }

        do {
            if let userId = userId {
                let response: MeResponse = try await NetworkManager.shared.request(
                    endpoint: "/api/users?id=\(userId)"
                )
                self.complaints = response.complaints
            } else {
                let meResponse: MeResponse = try await NetworkManager.shared.request(
                    endpoint: "/api/me"
                )

                self.complaints = meResponse.complaints
                self.user = meResponse.user

                let freshCache = LocalComplaintsCache(complaints: meResponse.complaints)
                await AppCacheStore.shared.save(freshCache, for: .complaints)
            }
        } catch {
            print("Error \(error)")

            if complaints.isEmpty {
                errMsg = "Не удалось загрузить данные"
            }
        }
    }
    
    func sendComplaint(targetId: Int, targetName: String, createdAt: String, content: String) async -> Bool {
        isLoading = true
        errMsg = nil
        
        defer { isLoading = false }
        
        do {
            let payload = AddComplaintPayload(targetId: targetId, targetName: targetName, content: content, createdAt: createdAt)
            let jsonData = try JSONEncoder().encode(payload)
            
            let _: EmptyResponse? = try await NetworkManager.shared.request(
                endpoint: "/api/complaint/add",
                method: "PATCH",
                body: jsonData
            )
            
            await fetchComplaints(for: targetId)
            return true
        } catch {
            print("Ошибка отправки жалобы: \(error)")
            errMsg = "Не удалось отправить жалобу. Попробуйте позже."
            return false
        }
    }

    func deleteComplaint(complaintId: Int) async -> Bool {
        isLoading = true
        errMsg = nil
        defer { isLoading = false }
        
        do {
            let _: EmptyResponse = try await NetworkManager.shared.request(endpoint: "/api/complaint/delete/\(complaintId)", method: "DELETE")
            
            if let index = complaints.firstIndex(where: { $0.id == complaintId }) {
                complaints.remove(at: index)
            }
            
            return true
        } catch {
            print("Error \(error)")
            errMsg = "Не удалось удалить заметку. Попробуйте позже"
            return false
        }
    }
    
//    private func apply(comments: [ComplaintModel]) {
//        complaints = comments
//
//        guard let first = comments.first else {
//            id = 0
//            targetName = ""
//            targetId = 0
//            authorName = ""
//            authorId = 0
//            content = ""
//            createdAt = ""
//            return
//        }
//
//        apply(complaint: first)
//    }
//
//    private func apply(complaint: ComplaintModel) {
//        id = complaint.id
//        targetName = complaint.targetName
//        targetId = complaint.targetId
//        authorName = complaint.authorName
//        authorId = complaint.authorId
//        content = complaint.content
//        createdAt = complaint.createdAt
//    }
}
