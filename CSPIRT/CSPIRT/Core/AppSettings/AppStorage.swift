import SwiftUI
import Combine

final class AppSettings: ObservableObject {
    @AppStorage("themeMode") var themeMode: String = "dark"
    var selectedColorScheme: ColorScheme? {
            switch themeMode {
            case "light":
                return .light
            case "dark":
                return .dark
            default:
                return nil // system
            }
        }
    @AppStorage("notificationsEnabled") var notificationsEnabled: Bool = true
    @AppStorage("hapticsEnabled") var hapticsEnabled: Bool = true
    @AppStorage("longCache") var longCache: String = "2 days"
}

