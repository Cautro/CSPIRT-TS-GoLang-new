import SwiftUI
import Combine

// MARK: - Модели данных
struct FallingItem: Identifiable {
    let id = UUID()
    var x: CGFloat
    var y: CGFloat
    var speed: CGFloat
    var rotation: Double
}

struct ShopItem: Identifiable {
    let id = UUID()
    let emoji: String
    let name: String
    let price: Int
}

struct LocationItem: Identifiable {
    let id = UUID()
    let name: String
    let price: Int
    let description: String
    let gradientColors: [Color]
}

struct Quest: Identifiable, Codable {
    let id: UUID
    let title: String
    let description: String
    let targetValue: Int
    var currentValue: Int
    let reward: Int
    let type: QuestType
    var isClaimed: Bool
    
    enum QuestType: String, Codable {
        case scoreInOneGame
        case totalCoinsEarned
        case gamesPlayed
        case totalItemsCaught
    }
}

// MARK: - Константы магазина
let availableShopItems: [ShopItem] = [
    ShopItem(emoji: "🥯", name: "Баранка", price: 0),
    ShopItem(emoji: "🍩", name: "Пончик", price: 50),
    ShopItem(emoji: "🍪", name: "Печенька", price: 150),
    ShopItem(emoji: "🍕", name: "Пицца", price: 300),
    ShopItem(emoji: "💎", name: "Алмаз", price: 1000)
]

let availableLocations: [LocationItem] = [
    LocationItem(name: "Школьная столовая", price: 0, description: "Где всё начиналось. Пахнет сладким чаем и булочками.", gradientColors: [Color.blue.opacity(0.2), Color.black.opacity(0.85)]),
    LocationItem(name: "Уютный Париж", price: 400, description: "Вид на Эйфелеву башню. Романтика и хруст свежей корочки.", gradientColors: [Color.orange.opacity(0.25), Color.purple.opacity(0.8)]),
    LocationItem(name: "Неоновый Токио", price: 1200, description: "Киберпанк-кондитерская под дождем. Баранки в стиле хай-тек.", gradientColors: [Color.purple.opacity(0.3), Color.init(white: 0.05)]),
    LocationItem(name: "Космическая Станция", price: 3000, description: "Выпечка в невесомости. Орбитальный масштаб ловли!", gradientColors: [Color.indigo.opacity(0.4), Color.black])
]

let globalQuestsPool: [Quest] = [
    Quest(id: UUID(), title: "Начинающий ловец", description: "Набери 15 очков за одну игру", targetValue: 15, currentValue: 0, reward: 30, type: .scoreInOneGame, isClaimed: false),
    Quest(id: UUID(), title: "Мастер комбо", description: "Набери 30 очков за одну игру", targetValue: 30, currentValue: 0, reward: 70, type: .scoreInOneGame, isClaimed: false),
    Quest(id: UUID(), title: "Легенда кондитерской", description: "Набери 50 очков за одну игру", targetValue: 50, currentValue: 0, reward: 150, type: .scoreInOneGame, isClaimed: false),
    Quest(id: UUID(), title: "Золотая лихорадка", description: "Заработай 25 монет", targetValue: 25, currentValue: 0, reward: 20, type: .totalCoinsEarned, isClaimed: false),
    Quest(id: UUID(), title: "Богач", description: "Заработай 100 монет", targetValue: 100, currentValue: 0, reward: 60, type: .totalCoinsEarned, isClaimed: false),
    Quest(id: UUID(), title: "Постоянный клиент", description: "Сыграй 3 игры", targetValue: 3, currentValue: 0, reward: 35, type: .gamesPlayed, isClaimed: false),
    Quest(id: UUID(), title: "Игроман", description: "Сыграй 7 игр", targetValue: 7, currentValue: 0, reward: 80, type: .gamesPlayed, isClaimed: false),
    Quest(id: UUID(), title: "Полная корзина", description: "Поймай 40 предметов", targetValue: 40, currentValue: 0, reward: 40, type: .totalItemsCaught, isClaimed: false),
    Quest(id: UUID(), title: "Голодный обжора", description: "Поймай 100 предметов", targetValue: 100, currentValue: 0, reward: 100, type: .totalItemsCaught, isClaimed: false)
]

// MARK: - Главное View игры
struct BarankaGameView: View {
    @AppStorage("bestScore") private var bestScore = 0
    @AppStorage("totalCoins") private var totalCoins = 0
    @AppStorage("selectedSkin") private var selectedSkin = "🥯"
    @AppStorage("unlockedSkins") private var unlockedSkinsStr = "🥯"
    
    // ЛОКАЦИИ
    @AppStorage("selectedLocation") private var selectedLocation = "Школьная столовая"
    @AppStorage("unlockedLocations") private var unlockedLocationsStr = "Школьная столовая"
    
    @AppStorage("dailyStreak") private var dailyStreak = 0
    @AppStorage("lastLoginDate") private var lastLoginDateStr = ""
    @AppStorage("streakClaimedToday") private var streakClaimedToday = false
    
    @AppStorage("activeQuestsJSON") private var activeQuestsJSON = ""
    @State private var activeQuests: [Quest] = []
    
    // ПРОКАЧКИ: ТИР 1 (Школьная столовая)
    @AppStorage("worker1_Count") private var worker1Count = 0
    @AppStorage("worker2_Count") private var worker2Count = 0
    @AppStorage("worker3_Count") private var worker3Count = 0
    
    // СЮЖЕТНЫЙ ПЕРЕХОД
    @AppStorage("isDreamUnlocked") private var isDreamUnlocked = false
    
    // ПРОКАЧКИ: ТИР 2 (Парижский Синдикат - открываются ПОСЛЕ покупки Мечты)
    @AppStorage("worker4_Count") private var worker4Count = 0
    @AppStorage("worker5_Count") private var worker5Count = 0
    @AppStorage("worker6_Count") private var worker6Count = 0
    
    @State private var coinAccumulator: Double = 0.0
    @State private var clickScale: CGFloat = 1.0
    @State private var selectedShopTab = 0
    
    private var unlockedSkins: [String] {
        unlockedSkinsStr.components(separatedBy: ",")
    }
    
    private var unlockedLocations: [String] {
        unlockedLocationsStr.components(separatedBy: ",")
    }
    
    // Стоимость Тир-1
    private var worker1Cost: Int { 50 + worker1Count * 35 }
    private var worker2Cost: Int { 300 + worker2Count * 180 }
    private var worker3Cost: Int { 1500 + worker3Count * 900 }
    private let secretDreamCost = 35000
    
    // Стоимость Тир-2 (Элитные сюжетные прокачки)
    private var worker4Cost: Int { 50000 + worker4Count * 35000 }
    private var worker5Cost: Int { 350000 + worker5Count * 220000 }
    private var worker6Cost: Int { 2500000 + worker6Count * 1800000 }
    
    // Общий расчет дохода в секунду с учетом обеих эр сюжета
    private var coinsPerSecond: Double {
        let tier1 = (Double(worker1Count) * 0.4) + (Double(worker2Count) * 2.5) + (Double(worker3Count) * 12.0)
        let tier2 = (Double(worker4Count) * 70.0) + (Double(worker5Count) * 450.0) + (Double(worker6Count) * 3000.0)
        return tier1 + tier2
    }
    
    @State private var score = 0
    @State private var sessionCoins = 0
    @State private var lives = 3
    @State private var isPlaying = false
    @State private var isGameOver = false
    @State private var showingShop = false
    @State private var showingQuests = false
    @State private var showingClicker = false
    @State private var showingStreakAlert = false
    
    @State private var playerX: CGFloat = 0
    private let playerWidth: CGFloat = 90
    private let playerHeight: CGFloat = 30
    
    @State private var fallingItems: [FallingItem] = []
    private let timer = Timer.publish(every: 0.016, on: .main, in: .common).autoconnect()
    
    @State private var spawnCounter = 0
    @State private var spawnInterval = 60

    var body: some View {
        GeometryReader { geometry in
            ZStack {
                backgroundLayer
                    .ignoresSafeArea()
                
                if showingClicker {
                    clickerScreenView
                } else if showingShop {
                    shopScreenView
                } else if showingQuests {
                    questsScreenView
                } else if !isPlaying && !isGameOver {
                    startScreenView
                } else if isGameOver {
                    gameOverScreenView
                } else {
                    gamePlayLayer(with: geometry)
                }
                
                if showingStreakAlert {
                    streakRewardAlertView
                }
            }
            .onAppear {
                playerX = geometry.size.width / 2
                checkDailyStreak()
                loadOrGenerateQuests()
            }
            .onReceive(timer) { _ in
                updateIdleIncome()
                guard isPlaying else { return }
                updateGameTick(in: geometry)
            }
        }
    }
}

// MARK: - Игровая логика
private extension BarankaGameView {
    func startGame() {
        score = 0
        sessionCoins = 0
        lives = 3
        fallingItems = []
        spawnCounter = 0
        spawnInterval = 60
        isGameOver = false
        isPlaying = true
        showingShop = false
        showingQuests = false
        showingClicker = false
    }
    
    func gameOver() {
        isPlaying = false
        isGameOver = true
        if score > bestScore {
            bestScore = score
        }
        updateQuestProgress(type: .gamesPlayed, value: 1)
        updateQuestProgress(type: .scoreInOneGame, value: score)
        saveQuests()
    }
    
    func updateIdleIncome() {
        let cps = coinsPerSecond
        if cps > 0 {
            coinAccumulator += cps * 0.016
            if coinAccumulator >= 1.0 {
                let integerCoins = Int(coinAccumulator)
                totalCoins += integerCoins
                updateQuestProgress(type: .totalCoinsEarned, value: integerCoins)
                coinAccumulator -= Double(integerCoins)
            }
        }
    }
    
    func updateGameTick(in geometry: GeometryProxy) {
        let groundLevel = geometry.size.height - 100
        let playerY = geometry.size.height - 110
        
        for id in fallingItems.indices {
            fallingItems[id].y += fallingItems[id].speed
            fallingItems[id].rotation += 2
        }
        
        var indicesToRemove: [Int] = []
        
        for (index, item) in fallingItems.enumerated() {
            if item.y >= (playerY - 20) && item.y <= (playerY + 20) {
                let hitBox = playerWidth / 2 + 15
                if abs(item.x - playerX) <= hitBox {
                    score += 1
                    sessionCoins += 1
                    totalCoins += 1
                    indicesToRemove.append(index)
                    
                    triggerHaptic(style: .light)
                    updateQuestProgress(type: .totalItemsCaught, value: 1)
                    updateQuestProgress(type: .totalCoinsEarned, value: 1)
                    
                    if score % 5 == 0 && spawnInterval > 25 {
                        spawnInterval -= 5
                    }
                    continue
                }
            }
            
            if item.y >= groundLevel {
                lives -= 1
                indicesToRemove.append(index)
                
                triggerHaptic(style: .heavy)
                if lives <= 0 {
                    gameOver()
                }
            }
        }
        
        for index in indicesToRemove.sorted(by: >) {
            if index < fallingItems.count {
                fallingItems.remove(at: index)
            }
        }
        
        spawnCounter += 1
        if spawnCounter >= spawnInterval {
            spawnCounter = 0
            let randomX = CGFloat.random(in: 40...(geometry.size.width - 40))
            let randomSpeed = CGFloat.random(in: 3.5...6.0) + CGFloat(score / 10)
            let newItem = FallingItem(x: randomX, y: -20, speed: randomSpeed, rotation: CGFloat.random(in: 0...360))
            fallingItems.append(newItem)
        }
    }
    
    func handleManualClick() {
        totalCoins += 1
        updateQuestProgress(type: .totalCoinsEarned, value: 1)
        triggerHaptic(style: .light)
        
        clickScale = 0.92
        withAnimation(.spring(response: 0.2, dampingFraction: 0.4)) {
            clickScale = 1.0
        }
    }
    
    func triggerHaptic(style: UIImpactFeedbackGenerator.FeedbackStyle) {
        let generator = UIImpactFeedbackGenerator(style: style)
        generator.impactOccurred()
    }
    
    func buyItem(_ item: ShopItem) {
        if totalCoins >= item.price && !unlockedSkins.contains(item.emoji) {
            totalCoins -= item.price
            unlockedSkinsStr += ",\(item.emoji)"
            selectedSkin = item.emoji
            triggerHaptic(style: .medium)
        }
    }
    
    func buyLocation(_ loc: LocationItem) {
        if totalCoins >= loc.price && !unlockedLocations.contains(loc.name) {
            totalCoins -= loc.price
            unlockedLocationsStr += ",\(loc.name)"
            selectedLocation = loc.name
            triggerHaptic(style: .medium)
        }
    }
    
    func checkDailyStreak() {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        let todayStr = formatter.string(from: Date())
        
        guard let yesterday = Calendar.current.date(byAdding: .day, value: -1, to: Date()) else { return }
        let yesterdayStr = formatter.string(from: yesterday)
        
        if lastLoginDateStr == todayStr {
            return
        } else if lastLoginDateStr == yesterdayStr {
            dailyStreak += 1
            streakClaimedToday = false
        } else {
            dailyStreak = 1
            streakClaimedToday = false
        }
        
        lastLoginDateStr = todayStr
        
        if !streakClaimedToday {
            let reward = dailyStreak * 15
            totalCoins += reward
            streakClaimedToday = true
            withAnimation {
                showingStreakAlert = true
            }
        }
    }
    
    func loadOrGenerateQuests() {
        if let data = activeQuestsJSON.data(using: .utf8),
           let decoded = try? JSONDecoder().decode([Quest].self, from: data) {
            self.activeQuests = decoded
            return
        }
        generateNewDailyQuests()
    }
    
    func generateNewDailyQuests() {
        let shuffled = globalQuestsPool.shuffled()
        let selectedQuests = Array(shuffled.prefix(3)).map { quest -> Quest in
            var q = quest
            q.currentValue = 0
            q.isClaimed = false
            return q
        }
        self.activeQuests = selectedQuests
        saveQuests()
    }
    
    func updateQuestProgress(type: Quest.QuestType, value: Int) {
        for i in 0..<activeQuests.count {
            if activeQuests[i].type == type && !activeQuests[i].isClaimed {
                if type == .scoreInOneGame {
                    activeQuests[i].currentValue = max(activeQuests[i].currentValue, value)
                } else {
                    activeQuests[i].currentValue += value
                }
                if activeQuests[i].currentValue > activeQuests[i].targetValue {
                    activeQuests[i].currentValue = activeQuests[i].targetValue
                }
            }
        }
        saveQuests()
    }
    
    func saveQuests() {
        if let encoded = try? JSONEncoder().encode(activeQuests),
           let jsonString = String(data: encoded, encoding: .utf8) {
            activeQuestsJSON = jsonString
        }
    }
    
    func claimQuestReward(at index: Int) {
        let quest = activeQuests[index]
        if quest.currentValue >= quest.targetValue && !quest.isClaimed {
            activeQuests[index].isClaimed = true
            totalCoins += quest.reward
            triggerHaptic(style: .medium)
            saveQuests()
        }
    }
}

// MARK: - Интерфейсы экранов
private extension BarankaGameView {
    var backgroundLayer: some View {
        ZStack {
            let currentLoc = availableLocations.first(where: { $0.name == selectedLocation }) ?? availableLocations[0]
            LinearGradient(
                colors: currentLoc.gradientColors,
                startPoint: .top,
                endPoint: .bottom
            )
            Color.black.opacity(0.15)
        }
    }
    
    func hudView() -> some View {
        HStack {
            HStack(spacing: 6) {
                Text(selectedSkin)
                Text("\(score)")
                    .font(.system(size: 22, weight: .bold, design: .rounded))
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 8)
            .background(Color.white.opacity(0.1))
            .cornerRadius(16)
            
            Spacer()
            
            HStack(spacing: 4) {
                ForEach(0..<3) { index in
                    Image(systemName: index < lives ? "heart.fill" : "heart")
                        .font(.system(size: 18))
                        .foregroundStyle(index < lives ? .red : .white.opacity(0.2))
                }
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 10)
            .background(Color.white.opacity(0.1))
            .cornerRadius(16)
        }
        .foregroundStyle(.white)
        .padding(.horizontal, 20)
        .padding(.top, 16)
    }
    
    func gamePlayLayer(with geometry: GeometryProxy) -> some View {
        ZStack(alignment: .top) {
            hudView()
            
            ForEach(fallingItems) { item in
                Text(selectedSkin)
                    .font(.system(size: 32))
                    .shadow(color: .white.opacity(0.3), radius: 5)
                    .rotationEffect(.degrees(item.rotation))
                    .position(x: item.x, y: item.y)
            }
            
            GlassGameCard {
                HStack {
                    Image(systemName: "arrow.left")
                    Spacer()
                    Text("Тарелка")
                        .font(.system(size: 13, weight: .bold, design: .rounded))
                        .textCase(.uppercase)
                        .tracking(1)
                    Spacer()
                    Image(systemName: "arrow.right")
                }
                .font(.system(size: 10, weight: .bold))
                .foregroundStyle(.white.opacity(0.8))
            }
            .frame(width: playerWidth, height: playerHeight)
            .position(x: playerX, y: geometry.size.height - 110)
            .gesture(
                DragGesture()
                    .onChanged { value in
                        let targetX = value.location.x
                        let minX = playerWidth / 2 + 10
                        let maxX = geometry.size.width - playerWidth / 2 - 10
                        playerX = max(minX, min(targetX, maxX))
                    }
            )
        }
    }
    
    var startScreenView: some View {
        GlassGameCard {
            VStack(spacing: 16) {
                HStack {
                    Text("🔥 \(dailyStreak) дн.")
                        .font(.system(size: 13, weight: .bold, design: .rounded))
                        .padding(.horizontal, 10)
                        .padding(.vertical, 4)
                        .background(Color.white.opacity(0.08))
                        .cornerRadius(12)
                    Spacer()
                    Text("📍 \(selectedLocation)")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundStyle(.white.opacity(0.7))
                    Spacer()
                    if coinsPerSecond > 0 {
                        Text(String(format: "⚙️ +%.1f/с", coinsPerSecond))
                            .font(.system(size: 13, weight: .medium, design: .rounded))
                            .foregroundStyle(.green)
                    }
                }
                
                Text(selectedSkin)
                    .font(.system(size: 70))
                    .bounceAnimation()
                
                VStack(spacing: 4) {
                    Text("Кэтч & Бейк")
                        .font(.system(size: 30, weight: .black, design: .rounded))
                        .foregroundStyle(.white)
                    
                    Text("Рекорд ловли: \(bestScore)")
                        .font(.system(size: 14, weight: .semibold, design: .rounded))
                        .foregroundStyle(.orange)
                    
                    Text("Баланс: 💰 \(totalCoins)")
                        .font(.system(size: 18, weight: .bold, design: .rounded))
                        .foregroundStyle(.yellow)
                }
                
                VStack(spacing: 8) {
                    Button {
                        startGame()
                    } label: {
                        Text("Ловить Баранки (Аркада)")
                            .font(.system(size: 16, weight: .bold, design: .rounded))
                            .foregroundStyle(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 12)
                            .background(Color.green)
                            .cornerRadius(14)
                    }
                    
                    Button {
                        showingClicker = true
                    } label: {
                        HStack {
                            Image(systemName: "hammer.fill")
                            Text("Пекарня-Кликер (\(isDreamUnlocked ? "Париж" : "Бизнес") )")
                        }
                        .font(.system(size: 16, weight: .bold, design: .rounded))
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 12)
                        .background(isDreamUnlocked ? Color.cyan : Color.orange)
                        .cornerRadius(14)
                    }
                    
                    HStack(spacing: 8) {
                        Button {
                            showingQuests = true
                        } label: {
                            HStack {
                                Image(systemName: "list.clipboard")
                                Text("Квесты")
                            }
                            .font(.system(size: 14, weight: .bold, design: .rounded))
                            .foregroundStyle(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 11)
                            .background(Color.blue.opacity(0.8))
                            .cornerRadius(12)
                        }
                        
                        Button {
                            showingShop = true
                        } label: {
                            HStack {
                                Image(systemName: "bag")
                                Text("Магазин")
                            }
                            .font(.system(size: 14, weight: .bold, design: .rounded))
                            .foregroundStyle(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 11)
                            .background(Color.purple.opacity(0.8))
                            .cornerRadius(12)
                        }
                    }
                }
            }
            .padding(20)
        }
        .padding(.horizontal, 24)
    }
    
    var clickerScreenView: some View {
        GlassGameCard {
            VStack(spacing: 16) {
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text(isDreamUnlocked ? "🥐 Глобальный Синдикат" : "🏪 Авто-Пекарня")
                            .font(.system(size: 18, weight: .bold, design: .rounded))
                            .foregroundStyle(isDreamUnlocked ? .cyan : .white)
                        Text(String(format: "Доход: +%.1f монет/с", coinsPerSecond))
                            .font(.system(size: 12))
                            .foregroundStyle(.green)
                    }
                    Spacer()
                    Button { showingClicker = false } label: {
                        Image(systemName: "xmark.circle.fill")
                            .font(.system(size: 24))
                            .foregroundStyle(.white.opacity(0.6))
                    }
                }
                
                ZStack {
                    Circle()
                        .fill(Color.white.opacity(0.04))
                        .frame(height: 110)
                    
                    Text(isDreamUnlocked ? "🥐" : selectedSkin)
                        .font(.system(size: 70))
                        .scaleEffect(clickScale)
                        .onTapGesture {
                            handleManualClick()
                        }
                }
                
                ScrollView(showsIndicators: false) {
                    VStack(spacing: 10) {
                        // --- ЭРА 1: ШКОЛЬНАЯ СТОЛОВАЯ ---
                        Text("Базовое производство")
                            .font(.system(size: 11, weight: .bold))
                            .foregroundStyle(.white.opacity(0.4))
                            .frame(maxWidth: .infinity, alignment: .leading)
                        
                        clickerWorkerRow(emoji: "🍩", name: "Пончик-Стажер", desc: "Приносит +0.4/с", count: worker1Count, cost: worker1Cost) {
                            if totalCoins >= worker1Cost {
                                totalCoins -= worker1Cost
                                worker1Count += 1
                                triggerHaptic(style: .medium)
                            }
                        }
                        
                        clickerWorkerRow(emoji: "🍪", name: "Печенька-Повар", desc: "Приносит +2.5/с", count: worker2Count, cost: worker2Cost) {
                            if totalCoins >= worker2Cost {
                                totalCoins -= worker2Cost
                                worker2Count += 1
                                triggerHaptic(style: .medium)
                            }
                        }
                        
                        clickerWorkerRow(emoji: "🍕", name: "Пицца-Директор", desc: "Приносит +12.0/с", count: worker3Count, cost: worker3Cost) {
                            if totalCoins >= worker3Cost {
                                totalCoins -= worker3Cost
                                worker3Count += 1
                                triggerHaptic(style: .medium)
                            }
                        }
                        
                        Divider().background(Color.white.opacity(0.2)).padding(.vertical, 4)
                        
                        // --- СЮЖЕТНАЯ СТЕНА (КВЕСТ МЕЧТЫ) ---
                        if !isDreamUnlocked {
                            VStack(spacing: 8) {
                                HStack {
                                    Text("✨").font(.system(size: 24))
                                    VStack(alignment: .leading) {
                                        Text("Главная мечта Баранки")
                                            .font(.system(size: 14, weight: .bold))
                                            .foregroundStyle(.cyan)
                                        Text("Билет в Париж и глазурная пластика")
                                            .font(.system(size: 11))
                                            .foregroundStyle(.white.opacity(0.6))
                                    }
                                    Spacer()
                                    Button {
                                        if totalCoins >= secretDreamCost {
                                            totalCoins -= secretDreamCost
                                            isDreamUnlocked = true
                                            triggerHaptic(style: .heavy)
                                        }
                                    } label: {
                                        Text("💰 \(secretDreamCost)")
                                            .font(.system(size: 12, weight: .bold))
                                            .foregroundStyle(.white)
                                            .padding(.horizontal, 10)
                                            .padding(.vertical, 6)
                                            .background(totalCoins >= secretDreamCost ? Color.cyan : Color.white.opacity(0.1))
                                            .cornerRadius(8)
                                    }
                                    .disabled(totalCoins < secretDreamCost)
                                }
                            }
                            .padding(10)
                            .background(Color.cyan.opacity(0.08))
                            .cornerRadius(12)
                        } else {
                            // --- ЭРА 2: СЮЖЕТНОЕ ПРОДОЛЖЕНИЕ (ОТКРЫВАЕТСЯ ПОСЛЕ ПОКУПКИ) ---
                            VStack(spacing: 8) {
                                Text("🇫🇷 СЮЖЕТ: ЭРА КРУАССАНА ОТКРЫТА!")
                                    .font(.system(size: 12, weight: .black, design: .rounded))
                                    .foregroundStyle(.yellow)
                                
                                Text("Баранка улетела в Париж, сделала пластику и превратилась в элитный Круассан на Монмартре! Но на этом история не кончается. Она выкупила Школьную Столовую и строит мировую империю!")
                                    .font(.system(size: 10))
                                    .foregroundStyle(.white.opacity(0.9))
                                    .multilineTextAlignment(.center)
                                    .padding(.bottom, 6)
                                
                                Text("Элитные сюжетные прокачки:")
                                    .font(.system(size: 11, weight: .bold))
                                    .foregroundStyle(.cyan)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                                
                                // НОВЫЕ ПРОКАЧКИ
                                clickerWorkerRow(emoji: "☕️", name: "Французский Бариста", desc: "Приносит +70.0/с", count: worker4Count, cost: worker4Cost) {
                                    if totalCoins >= worker4Cost {
                                        totalCoins -= worker4Cost
                                        worker4Count += 1
                                        triggerHaptic(style: .medium)
                                    }
                                }
                                
                                clickerWorkerRow(emoji: "⭐️", name: "Инспектор Мишлен", desc: "Приносит +450.0/с", count: worker5Count, cost: worker5Cost) {
                                    if totalCoins >= worker5Cost {
                                        totalCoins -= worker5Cost
                                        worker5Count += 1
                                        triggerHaptic(style: .medium)
                                    }
                                }
                                
                                clickerWorkerRow(emoji: "✈️", name: "Синдикат «Глазурный Путь»", desc: "Приносит +3000.0/с", count: worker6Count, cost: worker6Cost) {
                                    if totalCoins >= worker6Cost {
                                        totalCoins -= worker6Cost
                                        worker6Count += 1
                                        triggerHaptic(style: .heavy)
                                    }
                                }
                            }
                            .padding(10)
                            .background(Color.cyan.opacity(0.12))
                            .cornerRadius(12)
                            .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.cyan.opacity(0.3), lineWidth: 1))
                        }
                    }
                }
            }
            .padding(16)
        }
        .padding(.horizontal, 20)
    }
    
    func clickerWorkerRow(emoji: String, name: String, desc: String, count: Int, cost: Int, action: @escaping () -> Void) -> some View {
        HStack {
            Text(emoji).font(.system(size: 26))
            VStack(alignment: .leading, spacing: 2) {
                Text("\(name) (\(count) шт.)")
                    .font(.system(size: 13, weight: .bold))
                    .foregroundStyle(.white)
                Text(desc)
                    .font(.system(size: 10))
                    .foregroundStyle(.white.opacity(0.6))
            }
            Spacer()
            Button(action: action) {
                Text("💰 \(cost)")
                    .font(.system(size: 11, weight: .bold))
                    .foregroundStyle(.white)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 6)
                    .background(totalCoins >= cost ? Color.orange : Color.white.opacity(0.1))
                    .cornerRadius(8)
            }
            .disabled(totalCoins < cost)
        }
        .padding(8)
        .background(Color.white.opacity(0.05))
        .cornerRadius(10)
    }
    
    var shopScreenView: some View {
        GlassGameCard {
            VStack(spacing: 14) {
                HStack {
                    Text("Магазин")
                        .font(.system(size: 22, weight: .bold, design: .rounded))
                        .foregroundStyle(.white)
                    Spacer()
                    Text("💰 \(totalCoins)")
                        .foregroundStyle(.yellow)
                        .font(.system(size: 16, weight: .bold))
                    Spacer()
                    Button { showingShop = false } label: {
                        Image(systemName: "xmark.circle.fill").font(.system(size: 24)).foregroundStyle(.white.opacity(0.6))
                    }
                }
                
                HStack(spacing: 0) {
                    Button { selectedShopTab = 0 } label: {
                        Text("Скины").frame(maxWidth: .infinity).padding(.vertical, 8)
                            .background(selectedShopTab == 0 ? Color.purple.opacity(0.5) : Color.clear)
                    }
                    Button { selectedShopTab = 1 } label: {
                        Text("Локации").frame(maxWidth: .infinity).padding(.vertical, 8)
                            .background(selectedShopTab == 1 ? Color.purple.opacity(0.5) : Color.clear)
                    }
                }
                .foregroundStyle(.white)
                .font(.system(size: 14, weight: .bold))
                .background(Color.white.opacity(0.1))
                .cornerRadius(10)
                
                ScrollView(showsIndicators: false) {
                    VStack(spacing: 10) {
                        if selectedShopTab == 0 {
                            ForEach(availableShopItems) { item in
                                HStack {
                                    Text(item.emoji).font(.system(size: 30))
                                    Text(item.name).font(.system(size: 15, weight: .medium)).foregroundStyle(.white)
                                    Spacer()
                                    
                                    if unlockedSkins.contains(item.emoji) {
                                        Button {
                                            selectedSkin = item.emoji
                                        } label: {
                                            Text(selectedSkin == item.emoji ? "Выбран" : "Выбрать")
                                                .font(.system(size: 12, weight: .bold))
                                                .padding(.horizontal, 12)
                                                .padding(.vertical, 6)
                                                .background(selectedSkin == item.emoji ? Color.green : Color.blue)
                                                .cornerRadius(8)
                                        }
                                    } else {
                                        Button { buyItem(item) } label: {
                                            Text("💰 \(item.price)")
                                                .font(.system(size: 12, weight: .bold))
                                                .padding(.horizontal, 12)
                                                .padding(.vertical, 6)
                                                .background(totalCoins >= item.price ? Color.purple : Color.gray)
                                                .cornerRadius(8)
                                        }
                                        .disabled(totalCoins < item.price)
                                    }
                                }
                                .padding(8)
                                .background(Color.white.opacity(0.05))
                                .cornerRadius(10)
                            }
                        } else {
                            ForEach(availableLocations) { loc in
                                VStack(alignment: .leading, spacing: 4) {
                                    HStack {
                                        Text(loc.name).font(.system(size: 15, weight: .bold)).foregroundStyle(.white)
                                        Spacer()
                                        
                                        if unlockedLocations.contains(loc.name) {
                                            Button {
                                                selectedLocation = loc.name
                                            } label: {
                                                Text(selectedLocation == loc.name ? "Активна" : "Включить")
                                                    .font(.system(size: 12, weight: .bold))
                                                    .padding(.horizontal, 12)
                                                    .padding(.vertical, 6)
                                                    .background(selectedLocation == loc.name ? Color.green : Color.blue)
                                                    .cornerRadius(8)
                                            }
                                        } else {
                                            Button { buyLocation(loc) } label: {
                                                Text("💰 \(loc.price)")
                                                    .font(.system(size: 12, weight: .bold))
                                                    .padding(.horizontal, 12)
                                                    .padding(.vertical, 6)
                                                    .background(totalCoins >= loc.price ? Color.purple : Color.gray)
                                                    .cornerRadius(8)
                                            }
                                            .disabled(totalCoins < loc.price)
                                        }
                                    }
                                    Text(loc.description)
                                        .font(.system(size: 11))
                                        .foregroundStyle(.white.opacity(0.6))
                                        .multilineTextAlignment(.leading)
                                }
                                .padding(10)
                                .background(Color.white.opacity(0.05))
                                .cornerRadius(10)
                            }
                        }
                    }
                }
            }
            .padding(16)
        }
        .padding(.horizontal, 20)
    }
    
    var questsScreenView: some View {
        GlassGameCard {
            VStack(spacing: 14) {
                HStack {
                    Text("Ежедневные квесты")
                        .font(.system(size: 18, weight: .bold))
                        .foregroundStyle(.white)
                    Spacer()
                    Button { showingQuests = false } label: {
                        Image(systemName: "xmark.circle.fill").font(.system(size: 22)).foregroundStyle(.white.opacity(0.6))
                    }
                }
                
                VStack(spacing: 10) {
                    ForEach(0..<activeQuests.count, id: \.self) { i in
                        let quest = activeQuests[i]
                        VStack(alignment: .leading, spacing: 4) {
                            HStack {
                                Text(quest.title).font(.system(size: 14, weight: .bold)).foregroundStyle(.white)
                                Spacer()
                                Text("🎁 \(quest.reward) ц.")
                                    .font(.system(size: 12, weight: .bold))
                                    .foregroundStyle(.yellow)
                            }
                            Text(quest.description).font(.system(size: 11)).foregroundStyle(.white.opacity(0.6))
                            
                            HStack {
                                ProgressView(value: Double(quest.currentValue), total: Double(quest.targetValue))
                                    .tint(.green)
                                Text("\(quest.currentValue)/\(quest.targetValue)")
                                    .font(.system(size: 11, weight: .bold))
                                    .foregroundStyle(.white)
                                
                                if quest.isClaimed {
                                    Text("Взято").font(.system(size: 11)).foregroundStyle(.gray).padding(.leading, 4)
                                } else {
                                    Button { claimQuestReward(at: i) } label: {
                                        Text("Забрать")
                                            .font(.system(size: 11, weight: .bold))
                                            .padding(.horizontal, 8)
                                            .padding(.vertical, 4)
                                            .background(quest.currentValue >= quest.targetValue ? Color.green : Color.gray)
                                            .cornerRadius(6)
                                    }
                                    .disabled(quest.currentValue < quest.targetValue)
                                }
                            }
                        }
                        .padding(8)
                        .background(Color.white.opacity(0.06))
                        .cornerRadius(10)
                    }
                }
                
                Button { generateNewDailyQuests() } label: {
                    Text("Сбросить квесты (Изучить новые)")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundStyle(.white.opacity(0.7))
                        .padding(.top, 4)
                }
            }
            .padding(14)
        }
        .padding(.horizontal, 20)
    }
    
    var gameOverScreenView: some View {
        GlassGameCard {
            VStack(spacing: 20) {
                Text("Игра Окончена")
                    .font(.system(size: 26, weight: .black, design: .rounded))
                    .foregroundStyle(.red)
                
                VStack(spacing: 6) {
                    Text("Вы поймали: 🥯 \(score)")
                        .font(.system(size: 18, weight: .bold))
                    Text("Заработано монет: 💰 \(sessionCoins)")
                        .font(.system(size: 16))
                        .foregroundStyle(.yellow)
                }
                .foregroundStyle(.white)
                
                Button {
                    startGame()
                } label: {
                    Text("Играть Снова")
                        .font(.system(size: 16, weight: .bold))
                        .foregroundStyle(.white)
                        .padding(.vertical, 12)
                        .frame(maxWidth: .infinity)
                        .background(Color.green)
                        .cornerRadius(12)
                }
                
                Button {
                    isGameOver = false
                } label: {
                    Text("В Главное Меню")
                        .font(.system(size: 14))
                        .foregroundStyle(.white.opacity(0.6))
                }
            }
            .padding(20)
        }
        .padding(.horizontal, 30)
    }
    
    var streakRewardAlertView: some View {
        Color.black.opacity(0.5)
            .ignoresSafeArea()
            .overlay {
                GlassGameCard {
                    VStack(spacing: 16) {
                        Text("🎉 Ежедневный Бонус! 🎉")
                            .font(.system(size: 18, weight: .bold))
                            .foregroundStyle(.yellow)
                        Text("Вы заходите в игру \(dailyStreak) дней подряд!\nВаша награда:")
                            .font(.system(size: 13))
                            .multilineTextAlignment(.center)
                            .foregroundStyle(.white)
                        Text("💰 +\(dailyStreak * 15) монет")
                            .font(.system(size: 20, weight: .black))
                            .foregroundStyle(.green)
                        Button {
                            withAnimation { showingStreakAlert = false }
                        } label: {
                            Text("Отлично")
                                .font(.system(size: 14, weight: .bold))
                                .foregroundStyle(.white)
                                .padding(.horizontal, 24)
                                .padding(.vertical, 8)
                                .background(Color.blue)
                                .cornerRadius(10)
                        }
                    }
                    .padding(16)
                }
                .padding(.horizontal, 40)
            }
    }
}

// MARK: - Пользовательские UI Элементы
struct GlassGameCard<Content: View>: View {
    let content: Content
    init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }
    var body: some View {
        content
            .background(Color.white.opacity(0.08))
            .cornerRadius(24)
            .overlay(
                RoundedRectangle(cornerRadius: 24)
                    .stroke(Color.white.opacity(0.15), lineWidth: 1)
            )
            .shadow(color: .black.opacity(0.3), radius: 15, x: 0, y: 10)
    }
}

struct BounceModifier: ViewModifier {
    @State private var isAnimating = false
    func body(content: Content) -> some View {
        content
            .offset(y: isAnimating ? -6 : 6)
            .onAppear {
                withAnimation(.easeInOut(duration: 1.5).repeatForever(autoreverses: true)) {
                    isAnimating = true
                }
            }
    }
}

extension View {
    func bounceAnimation() -> some View {
        self.modifier(BounceModifier())
    }
}
