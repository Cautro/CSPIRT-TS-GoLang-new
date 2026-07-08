import Foundation
import SwiftUI
import UIKit
import Combine
import PhotosUI

extension UIImage {
    func resize(maxWidthHeight: CGFloat) -> UIImage? {
        let originalSize = size

        let widthRatio = maxWidthHeight / originalSize.width
        let heightRatio = maxWidthHeight / originalSize.height
        let ratio = min(widthRatio, heightRatio, 1.0)

        if ratio == 1.0 { return self }

        let newSize = CGSize(
            width: originalSize.width * ratio,
            height: originalSize.height * ratio
        )

        let renderer = UIGraphicsImageRenderer(size: newSize)
        return renderer.image { _ in
            self.draw(in: CGRect(origin: .zero, size: newSize))
        }
    }
}

@MainActor
final class ProfileViewModel: ObservableObject {
    
    @Published var userId: Int = 0
    
    @Published var Name: String = ""
    @Published var LastName: String = ""
    @Published var Login: String = ""
    @Published var Rating: Int = 0
    @Published var Avatar: String = ""
    @Published var Role: String = ""
    @Published var ClassName: String = ""
    @Published var ClassId: Int = 0
    @Published var FullName: [FullName] = []
    
    @Published var UserClass: ClassesModel?
    
    @Published var isLoading: Bool = false
    @Published var errMsg: String? = nil
    
    private let cacheTTL: TimeInterval = 2 * 24 * 60 * 60
    
    
    var avatarImage: UIImage? {
        guard !Avatar.isEmpty else { return nil }
        
        var cleanBase64 = Avatar
        if let range = cleanBase64.range(of: "base64,") {
            cleanBase64 = String(cleanBase64[range.upperBound...])
        }
        
        let trimmed = cleanBase64.trimmingCharacters(in: .whitespacesAndNewlines)
        
        guard let data = Data(base64Encoded: trimmed) else {
            print("Не удалось декодировать Base64 аватар")
            return nil
        }
        
        return UIImage(data: data)
    }
    
    func fetchProfile() async {
        errMsg = nil
        
        await loadCachedProfileIfPossible()
        
        if Name.isEmpty {
            isLoading = true
        }
        defer { isLoading = false }
        
        do {
            let response: MeResponse = try await NetworkManager.shared.request(
                endpoint: "/api/me"
            )
            
            let classRes: ClassesResponse = try await NetworkManager.shared.request(
                endpoint: "/api/classes?class_id=\(response.user.classId)"
            )
            self.UserClass = classRes.classes.first
            
            apply(user: response.user)
            await AppCacheStore.shared.save(response.user, for: .profile)
            
            let updatedSessionUser = CurrentUser(
                login: response.user.login,
                name: response.user.name,
                id: response.user.id ?? 0,
                role: response.user.role.rawValue,
                classId: response.user.classId,
                rating: response.user.rating
            )
            
            SessionManager.shared.handleSuccessfulLogin(user: updatedSessionUser)
            print("🍏 Сессия успешно синхронизирована с бэкендом. Новая роль: \(updatedSessionUser.role)")
            
        } catch {
            print("🔴 Ошибка загрузки профиля из сети: \(error.localizedDescription)")
            
            if Name.isEmpty {
                errMsg = "Не удалось загрузить профиль"
            }
        }
    }
        
    func updateRating(targetLogin: String, ratingChange: Int, reason: String) async -> Int? {
        errMsg = nil
        isLoading = true
        defer { isLoading = false }
        
        
        let body: [String: Any] = [
            "rating": ratingChange,
            "target_login": targetLogin,
            "reason": reason
        ]
        
        do {
            let jsonData = try JSONSerialization.data(withJSONObject: body)
            
            let res: UpdateRatingResponse = try await NetworkManager.shared.request(
                endpoint: "/api/rating/update",
                method: "PATCH",
                body: jsonData
            )
            
            print("🍏 Рейтинг успешно обновлен для @\(targetLogin)")
            return res.newRating
            
        } catch {
            print("🔴 Ошибка при обновлении рейтинга: \(error)")
            errMsg = "Ошибка обновления рейтинга"
            return nil
        }
    }
        
    func uploadAvatar(imageData: Data) async {
        guard userId != 0 else {
            errMsg = "Не удалось обновить аватар (неизвестен ID)"
            return
        }
        
        isLoading = true
        defer { isLoading = false }
        
        guard var uiImage = UIImage(data: imageData) else {
            errMsg = "Не удалось прочитать изображение"
            return
        }
        
        if let resizedImage = uiImage.resize(maxWidthHeight: 400) {
            uiImage = resizedImage
        }
        
        guard let compressedData = uiImage.jpegData(compressionQuality: 0.5) else {
            errMsg = "Не удалось сжать изображение"
            return
        }
        
        let base64String = "data:image/jpeg;base64," + compressedData.base64EncodedString()
        let body: [String: String] = ["avatar": base64String]
        
        do {
            let jsonData = try JSONSerialization.data(withJSONObject: body)
            
            let response: UpdateAvatarResponse = try await NetworkManager.shared.request(
                endpoint: "/api/user/update/avatar?id=\(userId)",
                method: "PATCH",
                body: jsonData
            )
            
            self.Avatar = response.Avatar ?? base64String
            await saveCurrentProfileToCache()
            
            print("🍏 Аватар успешно обновлён")
            
        } catch {
            print("🔴 Ошибка обновления аватара: \(error.localizedDescription)")
            errMsg = "Не удалось обновить аватар"
        }
    }
        
    func saveCurrentProfileToCache() async {
        let user = buildCurrentUserModel()
        await AppCacheStore.shared.save(user, for: .profile)
    }
    
    func loadCachedProfileIfPossible() async {
        guard let cachedUser = await AppCacheStore.shared.load(
            UserModel.self,
            for: .profile,
            maxAge: cacheTTL
        ) else {
            return
        }
        
        apply(user: cachedUser)
    }
    
    func apply(user: UserModel) {
        userId = user.id ?? 0
        Avatar = user.avatar?.value ?? ""
        Name = user.name
        LastName = user.lastName
        FullName = user.fullName
        Login = user.login
        Rating = user.rating
        Role = user.role.rawValue
        ClassName = user.className
        ClassId = user.classId
    }
    
    func buildCurrentUserModel() -> UserModel {
        UserModel(
            id: userId,
            avatar: Avatar.isEmpty ? nil : SQLNullString(string: Avatar, valid: true),
            name: Name,
            lastName: LastName,
            fullName: FullName,
            login: Login,
            rating: Rating,
            role: UserRole(rawValue: Role) ?? .user,
            className: ClassName,
            classId: ClassId
        )
    }
}
