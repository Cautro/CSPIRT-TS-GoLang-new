import Combine
import Foundation
import SwiftUI

@MainActor
final class MyClassViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errMsg: String?

    @Published var id: Int = 0
    @Published var name: String = ""
    @Published var grade: Int = 0
    @Published var letter: String = ""
    @Published var teacherLogin: String = ""
    @Published var teacher: UserModel?
    @Published var firstQuartyComplete: Int = 0
    @Published var secondQuartyComplete: Int = 0
    @Published var thirdQuertyComplete: Int = 0
    @Published var quarterComplete: Int = 0
    @Published var members: [UserModel] = []
    @Published var userTotalRating: Int = 0
    @Published var classTotalRating: Int = 0

    @Published var myClass: ClassesModel?

    private let cacheTTL: TimeInterval = 2 * 24 * 60 * 60 

    func fetchMyClass() async {
        errMsg = nil

        if let cache = await AppCacheStore.shared.load(
            LocalMyClassCache.self,
            for: .myClass,
            maxAge: cacheTTL
        ) {
            apply(myClass: cache.myClass)
        }

        if myClass == nil {
            isLoading = true
        }
        defer { isLoading = false }

        do {
            let meResponse: MeResponse = try await NetworkManager.shared.request(
                endpoint: "/api/me"
            )
            let classResponse: ClassesResponse = try await NetworkManager.shared.request(
                endpoint: "/api/classes?class_id=\(meResponse.user.classId)"
            )

            guard let classModel = classResponse.classes.first else {
                errMsg = "Класс не найден"
                return
            }

            apply(myClass: classModel)
            myClass = classModel

            let freshCache = LocalMyClassCache(myClass: classModel)
            await AppCacheStore.shared.save(freshCache, for: .myClass)

        } catch {
            print("error \(error)")

            if myClass == nil {
                errMsg = "Не удалось обновить данные"
            }
        }
    }

    private func apply(myClass: ClassesModel) {
        id = myClass.id
        name = myClass.name
        grade = myClass.grade
        letter = myClass.letter
        teacherLogin = myClass.teacherLogin
        teacher = myClass.teacher
        firstQuartyComplete = myClass.firstQuartyComplete
        secondQuartyComplete = myClass.secondQuartyComplete
        thirdQuertyComplete = myClass.thirdQuertyComplete
        quarterComplete = myClass.quarterComplete
        members = myClass.members
        userTotalRating = myClass.userTotalRating
        classTotalRating = myClass.classTotalRating
    }
}
