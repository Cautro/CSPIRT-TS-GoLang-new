import Foundation
import Combine

struct CurrentUser: Codable {
    let login: String
    let name: String
    let id: Int
    let role: String
    let classId: Int
    let rating: Int
}

@MainActor
final class SessionManager: ObservableObject {
    static let shared = SessionManager()
    
    private let authCookies = ["access_token", "refresh_token"]
    private let userDefaultsKey = "currentUserData"
    
    @Published var isAuthenticated: Bool = false
    @Published var currentUser: CurrentUser?
    
    private init() {
        checkSession()
    }
    
    func checkSession() {
        guard let allCookies = HTTPCookieStorage.shared.cookies else {
            setUnauthenticated()
            return
        }
        
        let hasAuthCookies = allCookies.contains { authCookies.contains($0.name) }
        
        if hasAuthCookies {
            self.currentUser = loadUserFromUserDefaults()
            self.isAuthenticated = true
            print("🔍 Сессия активна. Пользователь: \(self.currentUser?.login ?? "Неизвестен"), Роль: \(self.currentUser?.role ?? "Нет")")
        } else {
            setUnauthenticated()
        }
    }
    
    func handleSuccessfulLogin(user: CurrentUser) {
        saveUserToUserDefaults(user)
        self.currentUser = user
        self.isAuthenticated = true
    }
    
    func forceLogin(fallbackRole: String = "user") {
        let ghostUser = CurrentUser(login: "unknown", name: fallbackRole, id: 0, role: "unknown", classId: 0, rating: 0)
        handleSuccessfulLogin(user: ghostUser)
    }
    
    func logout() {
        if let allCookies = HTTPCookieStorage.shared.cookies {
            for cookie in allCookies {
                if cookie.domain.contains("cpirt.ru") || authCookies.contains(cookie.name) {
                    HTTPCookieStorage.shared.deleteCookie(cookie)
                }
            }
        }
        
        URLCache.shared.removeAllCachedResponses()
        
        UserDefaults.standard.removeObject(forKey: "accessToken")
        UserDefaults.standard.removeObject(forKey: userDefaultsKey)
        
        setUnauthenticated()
        print("🚪 Сессия закрыта, целевые куки и данные пользователя удалены.")
    }
    
    private func setUnauthenticated() {
        self.currentUser = nil
        self.isAuthenticated = false
    }
    
    private func saveUserToUserDefaults(_ user: CurrentUser) {
        if let encoded = try? JSONEncoder().encode(user) {
            UserDefaults.standard.set(encoded, forKey: userDefaultsKey)
        }
    }
    
    private func loadUserFromUserDefaults() -> CurrentUser? {
        guard let data = UserDefaults.standard.data(forKey: userDefaultsKey) else { return nil }
        return try? JSONDecoder().decode(CurrentUser.self, from: data)
    }
}
