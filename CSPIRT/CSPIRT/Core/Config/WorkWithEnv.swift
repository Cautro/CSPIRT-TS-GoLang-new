import Foundation

enum EnvironmentByEnv {
    static var baseURL: String {
        guard let urlString = Bundle.main.object(forInfoDictionaryKey: "API_BASE_URL") as? String else {
            fatalError("API_BASE_URL не найден в Info.plist. Проверь шаги настройки!")
        }
        return urlString
    }
}
