import Foundation

enum CacheKey: String, CaseIterable {
    case profile
    case dashboard
    case events
    case me
    case notes
    case myClass
    case complaints
    case parallels
    case schedule
}

struct CacheEntry<Value: Codable>: Codable {
    let savedAt: Date
    let value: Value
}

actor AppCacheStore {
    static let shared = AppCacheStore()

    private let memoryCache = NSCache<NSString, NSData>()
    private let fileManager = FileManager.default
    private let encoder = JSONEncoder()
    private let decoder = JSONDecoder()
    private let directory: URL

    init(folderName: String = "AppCache") {
        let baseURL = fileManager.urls(for: .cachesDirectory, in: .userDomainMask).first!
        directory = baseURL.appendingPathComponent(folderName, isDirectory: true)

        try? fileManager.createDirectory(
            at: directory,
            withIntermediateDirectories: true
        )
    }

    private func fileURL(for key: CacheKey) -> URL {
        directory.appendingPathComponent("\(key.rawValue).json")
    }
    
    func getCacheSizeInMB() async -> Double {
        do {
            let files = try fileManager.contentsOfDirectory(
                at: directory,
                includingPropertiesForKeys: [.fileSizeKey]
            )
            
            var totalBytes: Int = 0
            for file in files {
                let resources = try file.resourceValues(forKeys: [.fileSizeKey])
                if let fileSize = resources.fileSize {
                    totalBytes += fileSize
                }
            }
            
            return Double(totalBytes) / (1024.0 * 1024.0)
        } catch {
            print("❌ Ошибка при подсчете размера кеша: \(error)")
            return 0.0
        }
    }
    
    func hasCache() async -> Bool {
        do {
            let files = try fileManager.contentsOfDirectory(
                at: directory,
                includingPropertiesForKeys: nil
            )
            return !files.isEmpty
        } catch {
            return false
        }
    }

    func save<T: Codable>(_ value: T, for key: CacheKey) async {
        do {
            let entry = CacheEntry(savedAt: Date(), value: value)
            let data = try encoder.encode(entry)

            memoryCache.setObject(data as NSData, forKey: key.rawValue as NSString)
            try data.write(to: fileURL(for: key), options: [.atomic])
        } catch {
            print("❌ Cache save error for \(key.rawValue): \(error)")
        }
    }

    func load<T: Codable>(
        _ type: T.Type,
        for key: CacheKey,
        maxAge: TimeInterval? = nil
    ) async -> T? {
        let cacheKey = key.rawValue as NSString

        if let memoryData = memoryCache.object(forKey: cacheKey) as Data? {
            return decode(T.self, from: memoryData, maxAge: maxAge)
        }

        let url = fileURL(for: key)
        guard fileManager.fileExists(atPath: url.path) else { return nil }

        do {
            let data = try Data(contentsOf: url, options: [.mappedIfSafe])
            memoryCache.setObject(data as NSData, forKey: cacheKey)
            return decode(T.self, from: data, maxAge: maxAge)
        } catch {
            print("❌ Cache load error for \(key.rawValue): \(error)")
            return nil
        }
    }

    func remove(_ key: CacheKey) async {
        memoryCache.removeObject(forKey: key.rawValue as NSString)

        let url = fileURL(for: key)
        try? fileManager.removeItem(at: url)
    }

    func clearAll() async {
        memoryCache.removeAllObjects()

        do {
            let files = try fileManager.contentsOfDirectory(
                at: directory,
                includingPropertiesForKeys: nil
            )
            for file in files {
                try? fileManager.removeItem(at: file)
            }
        } catch {
            print("❌ Cache clear error: \(error)")
        }
    }

    private func decode<T: Codable>(
        _ type: T.Type,
        from data: Data,
        maxAge: TimeInterval?
    ) -> T? {
        do {
            let entry = try decoder.decode(CacheEntry<T>.self, from: data)

            if let maxAge {
                let age = Date().timeIntervalSince(entry.savedAt)
                if age > maxAge { return nil }
            }

            return entry.value
        } catch {
            return nil
        }
    }
}
