import SwiftUI

@main
struct CSPIRTApp: App {
    @StateObject private var sessionManager = SessionManager.shared
    @StateObject private var settings = AppSettings()
    
    var body: some Scene {
        WindowGroup {
            ZStack {
                if sessionManager.isAuthenticated {
                    MainTabView(classId: sessionManager.currentUser?.classId ?? 0)
                        .transition(.opacity)
                } else {
                    LoginView()
                        .transition(.opacity)
                }
            }
            .environmentObject(sessionManager)
            .animation(.easeInOut(duration: 0.3), value: sessionManager.isAuthenticated)
            .environmentObject(settings)
            .preferredColorScheme(settings.selectedColorScheme)
        }
    }
}
