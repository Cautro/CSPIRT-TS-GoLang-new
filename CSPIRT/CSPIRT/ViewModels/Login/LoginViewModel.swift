import SwiftUI
import Foundation
import Combine

struct LoginResponse: Codable {
    let accessToken: String
}

@MainActor
final class LoginViewModel: ObservableObject { 
    @Published var Login = ""
    @Published var Pass = ""
    @Published var isLoading = false
    @Published var errMsg: String?
    
    func performLogin() async {
            isLoading = true
            errMsg = nil
            
            let loginData = ["Login": Login, "Password": Pass]
            
            do {
                guard let jsonData = try? JSONEncoder().encode(loginData) else { return }
                
                let res: LoginResponse = try await NetworkManager.shared.request(
                    endpoint: "/login",
                    method: "POST",
                    body: jsonData
                )
                UserDefaults.standard.set(res.accessToken, forKey: "accessToken")
                SessionManager.shared.forceLogin()
                print("успех — рубильник переключен")
                
            } catch {
                print("🔴 ОШИБКА: \(error)")
                errMsg = "Ошибка авторизации"
            }
        
        isLoading = false
    }
}
