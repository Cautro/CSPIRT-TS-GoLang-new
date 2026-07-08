import Foundation
import Combine

@MainActor
final class MainPageViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errMsg: String?

    @Published var userName: String = ""
    @Published var me: UserModel?

    @Published var topClassUsers: [UserModel] = []
    @Published var availableEvents: [EventModel] = []
    @Published var latestNotes: [NoteModel] = []
    @Published var latestComplaints: [ComplaintModel] = []

    private let cacheTTL: TimeInterval = 2 * 24 * 60 * 60

    func loadDashboardData() async {
        errMsg = nil

        if let cached = await AppCacheStore.shared.load(
            LocalDashboardCache.self,
            for: .dashboard,
            maxAge: cacheTTL
        ) {
            apply(cache: cached)
        }

        if userName.isEmpty {
            isLoading = true
        }

        defer {
            isLoading = false
        }

        do {
            let meResponse: MeResponse = try await NetworkManager.shared.request(
                endpoint: "/api/me"
            )
            
            let updatedSessionUser = CurrentUser(
                login: meResponse.user.login,
                name: meResponse.user.name,
                id: meResponse.user.id ?? 0,
                role: meResponse.user.role.rawValue,
                classId: meResponse.user.classId,
                rating: meResponse.user.rating
            )
            
            SessionManager.shared.handleSuccessfulLogin(user: updatedSessionUser)
            
            let eventsResponse: [EventModel] = try await NetworkManager.shared.request(
                endpoint: "/api/events"
            )

            var sortedUsers: [UserModel] = []
            if meResponse.user.classId != 0 {
                do {
                    let classResponse: ClassUsersResponse = try await NetworkManager.shared.request(
                        endpoint: "/api/classes/\(meResponse.user.classId)/users"
                    )
                    sortedUsers = classResponse.users
                        .filter { $0.role != .admin && $0.role != .owner }
                        .sorted { $0.rating > $1.rating }
                } catch {
                    print("Ошибка загрузки списка класса: \(error)")
                }
            } else {
                print("У пользователя нет класса (Role: \(meResponse.user.role))")
            }

            let freshCache = LocalDashboardCache(
                userName: meResponse.user.name,
                me: meResponse.user,
                topClassUsers: Array(sortedUsers.prefix(5)),
                availableEvents: eventsResponse,
                latestNotes: Array(meResponse.notes.suffix(2).reversed()),
                latestComplaints: Array(meResponse.complaints.suffix(2).reversed())
            )

            apply(cache: freshCache)

            await AppCacheStore.shared.save(
                freshCache,
                for: .dashboard
            )

        } catch {
            print("Критическая ошибка загрузки главной страницы: \(error)")

            if userName.isEmpty {
                errMsg = "Не удалось обновить данные"
            }
        }
    }

    private func apply(cache: LocalDashboardCache) {
        userName = cache.userName
        me = cache.me
        topClassUsers = cache.topClassUsers
        availableEvents = cache.availableEvents
        latestNotes = cache.latestNotes
        latestComplaints = cache.latestComplaints
    }
}
