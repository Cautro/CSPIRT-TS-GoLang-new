import Combine
import Foundation

enum WeekDay: String, CaseIterable {
    case monday = "monday"
    case tuesday = "tuesday"
    case wednesday = "wednesday"
    case thursday = "thursday"
    case friday = "friday"
    case saturday = "saturday"
    
    var displayName: String {
        switch self {
        case .monday: return "Понедельник"
        case .tuesday: return "Вторник"
        case .wednesday: return "Среда"
        case .thursday: return "Четверг"
        case .friday: return "Пятница"
        case .saturday: return "Суббота"
        }
    }
    
    var shortName: String {
        switch self {
        case .monday: return "Пн"
        case .tuesday: return "Вт"
        case .wednesday: return "Ср"
        case .thursday: return "Чт"
        case .friday: return "Пт"
        case .saturday: return "Сб"
        }
    }
}

@MainActor
final class ScheduleViewModel: ObservableObject {
    private let cacheTTL: TimeInterval = 2 * 24 * 60 * 60
    
    @Published var isLoading: Bool = false
    @Published var errMsg: String? = nil
    
    @Published var schedules: [ScheduleModel] = []
    @Published var base: [ScheduleModel] = []
    @Published var current: [ScheduleModel] = []
    @Published var planned: [ScheduleModel] = []
    
    @Published var selectedDay: WeekDay = .monday
    
    var filteredCurrentSchedules: [ScheduleModel] {
        current.filter { $0.dayOfWeek.lowercased() == selectedDay.rawValue.lowercased() }
            .sorted { $0.lessonNumber < $1.lessonNumber }
    }
    
    init() {
        self.selectedDay = determineCurrentDay()
    }
    
    private func determineCurrentDay() -> WeekDay {
        let calendar = Calendar.current
        let weekdayIndex = calendar.component(.weekday, from: Date())
        
        switch weekdayIndex {
        case 2: return .monday
        case 3: return .tuesday
        case 4: return .wednesday
        case 5: return .thursday
        case 6: return .friday
        case 7: return .saturday
        case 1: return .monday
        default: return .monday
        }
    }
    
    func fetchSchedule() async {
        errMsg = nil
        
        if let cached = await AppCacheStore.shared.load(
            SchedulesResponse.self,
            for: .schedule,
            maxAge: cacheTTL
        ) {
            apply(cache: cached)
        }
        
        isLoading = true
        defer { isLoading = false }
        
        do {
            let meRes: MeResponse = try await NetworkManager.shared.request(endpoint: "/api/me")
            let scheduleRes: SchedulesResponse = try await NetworkManager.shared.request(endpoint: "/api/schedules?class_id=\(meRes.user.classId)&type=current")
            
            let freshCache = SchedulesResponse(
                schedules: scheduleRes.schedules,
                base: scheduleRes.base,
                current: scheduleRes.current,
                planned: scheduleRes.planned
            )
            apply(cache: freshCache)

            await AppCacheStore.shared.save(
                freshCache,
                for: .schedule
            )
        } catch {
            print("Error \(error)")
            errMsg = "Ошибка загрузки данных"
        }
    }
    
    private func apply(cache: SchedulesResponse){
        schedules = cache.schedules
        base = cache.base
        current = cache.current
        planned = cache.planned
    }
}
