import Foundation
import Combine
import SwiftUI

@MainActor
final class EventViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errMsg: String?
    
    @Published var featuredEvent: EventModel?
    
    @Published var availableEvents: [EventModel] = []
    @Published var pastEvents: [EventModel] = []
    
    @Published var classesInEvents: [ClassesModel] = []
    @Published var allUsers: [SafeUserModel] = []
    
    @Published var isOwner: Bool = false
    
    private let cacheTTL: TimeInterval = 2 * 24 * 60 * 60
    
    func fetchEvents() async {
        if let cache = await AppCacheStore.shared.load(
            LocalEventsCache.self,
            for: .events,
            maxAge: cacheTTL
        ) {
            apply(cache: cache)
        }
        
        if availableEvents.isEmpty && pastEvents.isEmpty {
            isLoading = true
        }
        
        defer { self.isLoading = false }
        
        do {
            let meRes: MeResponse = try await NetworkManager.shared.request(endpoint: "/api/me")
            self.isOwner = (meRes.user.role == UserRole.owner)
            
            async let availableReq: [EventModel] = NetworkManager.shared.request(endpoint: "/api/events\(self.isOwner ? "" : "?class=\(meRes.user.classId)")")
            async let pastReq: [EventModel] = NetworkManager.shared.request(endpoint: "/api/events")
            async let classesReq: ClassesResponse = NetworkManager.shared.request(endpoint: "/api/classes")
            
            let (fetchedAvailable, fetchedPast, fetchedClasses) = try await (availableReq, pastReq, classesReq)
            
            let activeEvents = fetchedAvailable.filter { $0.status != "completed" }
            self.availableEvents = activeEvents
            self.featuredEvent = activeEvents.first
            self.pastEvents = fetchedPast.filter { $0.status == "completed" }
            
            self.classesInEvents = fetchedClasses.classes

            let freshCache = LocalEventsCache(
                availableEvents: activeEvents,
                pastEvents: self.pastEvents
            )
            
            await AppCacheStore.shared.save(
                freshCache,
                for: .events
            )
            
            self.errMsg = nil
            
        } catch {
            print("Failed to load events: \(error)")
            if availableEvents.isEmpty && pastEvents.isEmpty {
                self.errMsg = "Не удалось загрузить данные"
            }
        }
    }
    
    private func apply(cache: LocalEventsCache) {
        availableEvents = cache.availableEvents
        pastEvents = cache.pastEvents
        featuredEvent = cache.availableEvents.first
    }
    
    // MARK: - Управление мероприятиями (Только для Owner)
    
    func createEvent(title: String, description: String, reward: Int, startedAt: String, classes: [Int]) async throws {
        let payload = EventPayload(
            title: title,
            status: "scheduled",
            ratingReward: reward,
            description: description,
            startedAt: startedAt,
            players: [],
            classes: classes
        )
        let _: EmptyResponse = try await NetworkManager.shared.requestWithBody(
            endpoint: "/api/event/add",
            method: "PATCH",
            body: payload
        )
        await fetchEvents()
    }
    
    func deleteEvent(eventId: Int) async throws {
        let _: EmptyResponse = try await NetworkManager.shared.requestWithoutBody(
            endpoint: "/api/event/delete/\(eventId)",
            method: "DELETE"
        )
        await fetchEvents()
    }
    
    func updateEvent(eventId: Int, title: String, description: String, reward: Int, status: String, startedAt: String, classes: [Int], players: [Int]) async throws {
        let payload = EventPayload(
            title: title,
            status: status,
            ratingReward: reward,
            description: description,
            startedAt: startedAt,
            players: players,
            classes: classes
        )
        let _: EmptyResponse = try await NetworkManager.shared.requestWithBody(
            endpoint: "/api/event/\(eventId)/update",
            method: "PATCH",
            body: payload
        )
        await fetchEvents()
    }
    
    func completeEvent(eventId: Int, ratingReward: Int, classReward: Int) async throws {
        let payload = EventCompletePayload(ratingReward: ratingReward, classReward: classReward)
        let _: EmptyResponse = try await NetworkManager.shared.requestWithBody(
            endpoint: "/api/event/\(eventId)/complete",
            method: "PATCH",
            body: payload
        )
        await fetchEvents()
    }
    
    // MARK: - Управление участниками
    
    func getEventPlayers(for event: EventModel) -> [SafeUserModel] {
        return allUsers.filter { event.players.contains($0.id) }
    }
    
    func addPlayersToEvent(eventId: Int, playerIds: [Int]) async throws {
        let payload = EventPlayersPayload(playerIds: playerIds)
        let _: EmptyResponse = try await NetworkManager.shared.requestWithBody(
            endpoint: "/api/event/\(eventId)/players/add",
            method: "PATCH",
            body: payload
        )
        await fetchEvents()
    }
    
    func removePlayersFromEvent(eventId: Int, playerIds: [Int]) async throws {
        let payload = EventPlayersPayload(playerIds: playerIds)
        let _: EmptyResponse = try await NetworkManager.shared.requestWithBody(
            endpoint: "/api/event/\(eventId)/players/delete",
            method: "DELETE",
            body: payload
        )
        await fetchEvents()
    }
    
    func fetchAllUsers() async {
        do {
            let users: [SafeUserModel] = try await NetworkManager.shared.request(endpoint: "/api/users")
            self.allUsers = users
        } catch {
            print("Failed to fetch all users: \(error)")
        }
    }
}

struct LocalEventsCache: Codable {
    let availableEvents: [EventModel]
    let pastEvents: [EventModel]
}
