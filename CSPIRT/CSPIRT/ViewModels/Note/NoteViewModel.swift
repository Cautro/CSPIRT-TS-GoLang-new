import Foundation
import Combine

@MainActor
final class NoteViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errMsg: String?

    @Published var id: Int = 0
    @Published var targetName: String = ""
    @Published var targetId: Int = 0
    @Published var authorName: String = ""
    @Published var authorId: Int = 0
    @Published var content: String = ""
    @Published var createdAt: String = ""

    @Published var notes: [NoteModel] = []

    private let cacheTTL: TimeInterval = 2 * 24 * 60 * 60

    func fetchNotes(for userId: Int? = nil) async {
        errMsg = nil

        if userId == nil {
            if let cache = await AppCacheStore.shared.load(
                LocalNoteCache.self,
                for: .notes,
                maxAge: cacheTTL
            ) {}
        }

        if notes.isEmpty {
            isLoading = true
        }
        defer { isLoading = false }

        do {
            if let userId = userId {
                let response: MeResponse = try await NetworkManager.shared.request(
                    endpoint: "/api/users?id=\(userId)"
                )
                self.notes = response.notes
            } else {
                let meResponse: MeResponse = try await NetworkManager.shared.request(
                    endpoint: "/api/me"
                )
                
                let freshNotes = meResponse.notes
                notes = freshNotes
                
                let freshCache = LocalNoteCache(notes: freshNotes)
                await AppCacheStore.shared.save(freshCache, for: .notes)
            }
        } catch {
            print("Error \(error)")

            if notes.isEmpty {
                errMsg = "Не удалось загрузить данные"
            }
        }
    }

    func sendNote(targetId: Int, targetName: String, createdAt: String, content: String) async -> Bool {
        isLoading = true
        errMsg = nil
        
        defer { isLoading = false }
        
        do {
            let payload = AddNotePayload(
                targetId: targetId,
                targetName: targetName,
                content: content,
                createdAt: createdAt
            )
            let jsonData = try JSONEncoder().encode(payload)
            
            let _: EmptyResponse? = try await NetworkManager.shared.request(
                endpoint: "/api/note/add",
                method: "PATCH",
                body: jsonData
            )
            
            await fetchNotes(for: targetId)
            return true
        } catch {
            print("Ошибка отправки заметки: \(error)")
            errMsg = "Не удалось сохранить заметку. Попробуйте позже."
            return false
        }
    }
    
    func deleteNote(noteId: Int) async -> Bool {
        isLoading = true
        errMsg = nil
        defer { isLoading = false }
        
        do {
            let deleteRes: EmptyResponse = try await NetworkManager.shared.request(endpoint: "/api/note/delete/\(noteId)", method: "DELETE")
            
            if let index = notes.firstIndex(where: { $0.id == noteId }) {
                notes.remove(at: index)
            }
            
            return true
        } catch {
            print("Error \(error)")
            errMsg = "Не удалось удалить заметку. Попробуйте позже"
            return false
        }
    }
    
//    private func apply(notes: [NoteModel]) {
//        self.notes = notes
//
//        guard let first = notes.first else {
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
//        apply(note: first)
//    }
//
//    private func apply(note: NoteModel) {
//        id = note.id
//        targetName = note.targetName
//        targetId = note.targetId
//        authorName = note.authorName
//        authorId = note.authorId
//        content = note.content
//        createdAt = note.createdAt
//    }
}
