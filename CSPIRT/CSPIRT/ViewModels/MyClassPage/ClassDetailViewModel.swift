import Foundation
import Combine

@MainActor
final class ClassDetailViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errMsg: String?

    @Published var id: Int = 0
    @Published var name: String = ""
    @Published var grade: Int = 0
    @Published var letter: String = ""
    @Published var teacher: UserModel?
    @Published var members: [UserModel] = []
    @Published var userTotalRating: Int = 0
    
    private let cacheTTL: TimeInterval = 24 * 60 * 60
    let classId: Int

    init(classId: Int) {
        self.classId = classId
    }

    func fetchClassDetails() async {
        errMsg = nil

        isLoading = true
        defer { isLoading = false }

        do {
            let classResponse: ClassesResponse = try await NetworkManager.shared.request(
                endpoint: "/api/classes?class_id=\(classId)"
            )

            guard let classModel = classResponse.classes.first else {
                errMsg = "Класс не найден"
                return
            }

            apply(myClass: classModel)

        } catch {
            print("error \(error)")
            errMsg = "Не удалось обновить данные"
        }
    }

    private func apply(myClass: ClassesModel) {
        id = myClass.id
        name = myClass.name
        grade = myClass.grade
        letter = myClass.letter
        teacher = myClass.teacher
        members = myClass.members
        userTotalRating = myClass.userTotalRating
    }
}
