import Foundation

enum NetworkError: Error {
    case unauthorized
    case badResponse
    case decodingError
    case serverError(String)
    
    var errorDescription: String? {
        switch self {
        case .unauthorized:
            return "Ошибка: Не авторизован (401)"
        case .badResponse:
            return "Ошибка: Плохой ответ от сервера"
        case .decodingError:
            return "Ошибка: Не удалось распарсить JSON"
        case .serverError(let message):
            return message
        }
    }
}

final class NetworkManager {
    static let shared = NetworkManager()
    
    private let session: URLSession
    
    private init() {
        let config = URLSessionConfiguration.default
        config.httpCookieStorage = .shared
        config.httpShouldSetCookies = true
        config.timeoutIntervalForRequest = 30
        self.session = URLSession(configuration: config)
    }
    
    func request<T: Decodable>(
        endpoint: String,
        method: String = "GET",
        body: Data? = nil
    ) async throws -> T {
        
        let url = URL(string: EnvironmentByEnv.baseURL + endpoint)!
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = method
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        urlRequest.httpBody = body
        
        if let token = UserDefaults.standard.string(forKey: "accessToken") {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        let (data, response) = try await session.data(for: urlRequest)
        
        if let httpResponse = response as? HTTPURLResponse {
            print("Запрос к: \(endpoint) | Статус: \(httpResponse.statusCode)")
            
            if let jsonString = String(data: data, encoding: .utf8) {
                print("Ответ сервера (сырой JSON): \(jsonString)")
            }
            
            if let headerFields = httpResponse.allHeaderFields as? [String: String],
               let url = httpResponse.url {
                let cookies = HTTPCookie.cookies(withResponseHeaderFields: headerFields, for: url)
                
                HTTPCookieStorage.shared.setCookies(cookies, for: url, mainDocumentURL: nil)
//                
//                print("🍪 Пришло кук от сервера: \(cookies.count)")
//                for cookie in cookies {
//                    print("   - Имя: \(cookie.name)")
//                    print("   - Домен: \(cookie.domain)")
//                    print("   - Путь: \(cookie.path)")
//                    print("   - Secure: \(cookie.isSecure)")
//                    print("   - Expires: \(String(describing: cookie.expiresDate))")
//                }
            }
        }
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw NetworkError.badResponse
        }
        
        if httpResponse.statusCode == 401 {
            if endpoint.contains("login") {
                print("401 на эндпоинте логина. Просто выбрасываем ошибку.")
                throw NetworkError.unauthorized
            }
            
            print("401: пробуем рефреш...")
            
            let success = await refreshAccessToken()
            
            if success {
                print("Рефреш успешен, повторяем запрос.")
                return try await self.request(endpoint: endpoint, method: method, body: body)
            } else {
                SessionManager.shared.logout()
                throw NetworkError.unauthorized
            }
        }
        
        guard (200...299).contains(httpResponse.statusCode) else {
            throw NetworkError.serverError("Код ошибки: \(httpResponse.statusCode)")
        }
        
        do {
            return try JSONDecoder().decode(T.self, from: data)
        } catch {
            print("Ошибка декодирования: \(error)")
            throw NetworkError.decodingError
        }
    }
    
    private func refreshAccessToken() async -> Bool {
            guard let url = URL(string: EnvironmentByEnv.baseURL + "/api/refresh") else { return false }
            
            var request = URLRequest(url: url)
            request.httpMethod = "POST"
            
            do {
                let (data, response) = try await session.data(for: request)
                
                if let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 {
                    struct RefreshResponse: Decodable {
                        let accessToken: String
                    }
                    
                    do {
                        let refreshData = try JSONDecoder().decode(RefreshResponse.self, from: data)
                        UserDefaults.standard.set(refreshData.accessToken, forKey: "accessToken")
                        print("Рефреш успешен")
                        return true
                    } catch {
                        print("Ошибка: Сервер вернул 200 на рефреш, но распарсить accessToken не удалось. \(error)")
                        return false
                    }
                    
                }
                
                print("Рефреш отклонен сервером. Статус: \((response as? HTTPURLResponse)?.statusCode ?? 0)")
                return false
                
            } catch {
                print("Ошибка сети при рефреше: \(error)")
                return false
            }
        }
}

extension NetworkManager {
    func requestWithBody<T: Encodable, U: Decodable>(endpoint: String, method: String, body: T?) async throws -> U {
        
        let baseURLString = UserDefaults.standard.string(forKey: "baseURL") ?? "https://cpirt.ru/backend"
        
        guard let url = URL(string: baseURLString + endpoint) else {
            throw URLError(.badURL)
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        if let token = UserDefaults.standard.string(forKey: "jwt_token") {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        if let body = body {
            request.httpBody = try JSONEncoder().encode(body)
        }
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw URLError(.badServerResponse)
        }
        
        guard (200...299).contains(httpResponse.statusCode) else {
            print("API Error: \(httpResponse.statusCode) - \(String(data: data, encoding: .utf8) ?? "")")
            throw URLError(.badServerResponse)
        }
        
        if U.self == EmptyResponse.self {
            return EmptyResponse() as! U
        }
        
        return try JSONDecoder().decode(U.self, from: data)
    }
    
    func requestWithoutBody<U: Decodable>(endpoint: String, method: String) async throws -> U {
        let dummyBody: [String: String]? = nil
        return try await requestWithBody(endpoint: endpoint, method: method, body: dummyBody)
    }
}
