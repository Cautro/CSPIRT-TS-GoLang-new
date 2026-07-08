import Foundation
import Combine

@MainActor
final class ParallelsViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errMsg: String?
    
    @Published var parallels: [ParallelModel] = []
    @Published var bestClass: ClassesModel?
    @Published var classesInParallel: [ClassesModel] = []
    @Published var me: MeResponse?
    
    @Published var selectedParallel: ParallelModel?
    @Published var selectedParallelClasses: [ClassesModel] = []
    @Published var selectedBestClass: ClassesModel?
    @Published var selectedTescher: UserModel?
    @Published var isShowingParallelDetails = false
    
    private let cacheTTL: TimeInterval = 2 * 24 * 60 * 60
    
    func fetchParallels() async {
        errMsg = nil
        if let cache = await AppCacheStore.shared.load(
            LocalParallelCache.self,
            for: .parallels,
            maxAge: cacheTTL
        ) {
            apply(cache: cache)
        }
        
        if parallels.isEmpty {
            print("Array is empty")
            self.isLoading = true
        }

        defer { self.isLoading = false }
        
        do {
            let meResponse: MeResponse = try await NetworkManager.shared.request(endpoint: "/api/me")
            let parallelsRes: ParallelResponse = try await NetworkManager.shared.request(endpoint: "/api/classes/parallel")
            
            self.parallels = parallelsRes.parallels
            
            guard let firstParallel = parallelsRes.parallels.first else {
                self.parallels = []
                return
            }
            
            let bestClassResponse: ClassesModel =
                try await NetworkManager.shared.request(
                    endpoint: "/api/classes?class_id=\(firstParallel.bestClassId ?? 0)"
                )

            self.bestClass = bestClassResponse
            
            let classesRes: [ClassesModel] = try await NetworkManager.shared.request(endpoint: "/api/classes/parallel/\(firstParallel.id)/classes")
            
            self.classesInParallel = classesRes
            self.me = meResponse
            
            let cacheToSave = LocalParallelCache(
                parallels: self.parallels,
                bestClass: self.bestClass,
                classes: self.classesInParallel,
            )
            
            await AppCacheStore.shared.save(cacheToSave, for: .parallels)
            
        } catch {
            print("Failed to load parallels: \(error)")
            errMsg = "Не удалось загрузить данные"
        }
    }
    
    func selectParallel(_ parallel: ParallelModel) async {
        selectedParallel = parallel
        isShowingParallelDetails = true
        errMsg = nil

        do {
            let classesRes: ClassesResponse = try await NetworkManager.shared.request(
                endpoint: "/api/classes/parallel/\(parallel.id)/classes"
            )
            
            let sortedClasses = classesRes.classes.sorted {
                ($0.userTotalRating + $0.classTotalRating) > ($1.userTotalRating + $1.classTotalRating)
            }
            
            selectedParallelClasses = sortedClasses
            
            if let bestId = parallel.bestClassId,
               let exactBestClass = sortedClasses.first(where: { $0.id == bestId }) {
                selectedBestClass = exactBestClass
            } else {
                selectedBestClass = sortedClasses.first
            }

        } catch {
            print("Failed to load parallel details: \(error)")
            errMsg = "Не удалось загрузить классы параллели"
        }
    }

    
    private func apply(cache: LocalParallelCache) {
        parallels = cache.parallels
        bestClass = cache.bestClass
        classesInParallel = cache.classes
    }
}
