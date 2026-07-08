import Foundation

enum MockDataSeeder {
    static func seedIfNeeded() async {
        #if DEBUG
        let defaults = UserDefaults.standard
        let key = "has_seeded_mock_data"

        guard !defaults.bool(forKey: key) else { return }

        // Профиль
        let mockUser = UserModel(
            id: 1,
            avatar: nil,
            name: "Артём",
            lastName: "Иванов",
            fullName: [
                FullName(Name: "Артём", LastName: "Иванов", MiddleName: "")
            ],
            login: "artem",
            rating: 120,
            role: .user,
            className: "10A",
            classId: 1
        )

        // Заметки
        let mockNotes: [NoteModel] = [
            NoteModel(
                id: 1,
                targetName: "Артём Иванов",
                targetId: 1,
                authorName: "Классный руководитель",
                authorId: 2,
                content: "Подготовить проект по информатике.",
                createdAt: "2026-06-19T10:00:00Z"
            )
        ]

        // Жалобы
        let mockComplaints: [ComplaintModel] = [
            ComplaintModel(
                id: 1,
                targetName: "Артём Иванов",
                targetId: 1,
                authorName: "Учитель",
                authorId: 2,
                content: "Опоздание на урок.",
                createdAt: "2026-06-19T09:30:00Z"
            )
        ]

        // События
        let mockEvents: [EventModel] = [
            EventModel(
                id: 1,
                title: "Олимпиада по математике",
                status: "Прими участие и получи рейтинг.",
                baseRatingReward: 25,
                description: "2026-04-19T12:00:00Z",
                createdAt: "2026-06-20T12:00:00Z",
                startedAt: "active",
                players: [],
                classes: [1]
            )
        ]

        // Класс
        let mockClass = ClassesModel(
            id: 1,
            name: "10A",
            grade: 10,
            letter: "A",
            teacherLogin: "teacher1",
            teacher: mockUser,
            firstQuartyComplete: 0,
            secondQuartyComplete: 0,
            thirdQuertyComplete: 0,
            quarterComplete: 0,
            members: [mockUser],
            userTotalRating: 120,
            classTotalRating: 120
        )

        // Параллели
        let mockParallels: [ParallelModel] = [
            ParallelModel(
                id: 10,
                name: "10 параллель",
                bestClassId: 1,
                classesIds: [1],
                classTotalRating: 0
            )
        ]

        await AppCacheStore.shared.save(mockUser, for: .profile)
        await AppCacheStore.shared.save(
            LocalDashboardCache(
                userName: mockUser.name,
                me: mockUser,
                topClassUsers: [mockUser],
                availableEvents: mockEvents,
                latestNotes: mockNotes,
                latestComplaints: mockComplaints
            ),
            for: .dashboard
        )

        await AppCacheStore.shared.save(
            LocalMyClassCache(myClass: mockClass),
            for: .myClass
        )

        await AppCacheStore.shared.save(
            LocalNoteCache(notes: mockNotes),
            for: .notes
        )

        await AppCacheStore.shared.save(
            LocalComplaintsCache(complaints: mockComplaints),
            for: .complaints
        )

        await AppCacheStore.shared.save(
            LocalEventsCache(
                availableEvents: mockEvents,
                pastEvents: []
            ),
            for: .events
        )

        await AppCacheStore.shared.save(
            ParallelResponse(parallels: mockParallels),
            for: .parallels
        )

        defaults.set(true, forKey: key)
        #endif
    }
}
