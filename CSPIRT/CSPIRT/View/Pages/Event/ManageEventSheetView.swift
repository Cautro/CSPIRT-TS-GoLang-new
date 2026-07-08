import SwiftUI

struct ManageEventSheet: View {
    @ObservedObject var viewModel: EventViewModel
    let event: EventModel
    @Environment(\.dismiss) private var dismiss
    
    @State private var selectedTab = 0
    
    // States for Editing
    @State private var title: String
    @State private var description: String
    @State private var rewardString: String
    @State private var status: String
    @State private var startDate: Date
    @State private var selectedClasses: Set<Int>
    
    // States for Players
    @State private var eventPlayers: [SafeUserModel] = []
    @State private var showAddPlayerSheet = false
    @State private var isLoadingPlayers = false
    
    // States for Completion
    @State private var ratingRewardString: String
    @State private var classRewardString: String = "0"
    
    @State private var isSubmitting = false
    @State private var errorMessage: String?
    
    let statuses = ["scheduled", "active", "completed"]
    
    init(viewModel: EventViewModel, event: EventModel) {
        self.viewModel = viewModel
        self.event = event
        
        _title = State(initialValue: event.title)
        _description = State(initialValue: event.description)
        _rewardString = State(initialValue: String(event.baseRatingReward))
        _status = State(initialValue: event.status)
        _selectedClasses = State(initialValue: Set(event.classes))
        
        _ratingRewardString = State(initialValue: String(event.baseRatingReward))
        
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = formatter.date(from: event.startedAt) ?? ISO8601DateFormatter().date(from: event.startedAt) {
            _startDate = State(initialValue: date)
        } else {
            _startDate = State(initialValue: Date())
        }
    }
    
    var body: some View {
        NavigationView {
            ZStack {
                Color(red: 0.05, green: 0.05, blue: 0.07).ignoresSafeArea()
                
                VStack(spacing: 0) {
                    Picker("Управление", selection: $selectedTab) {
                        Text("Настройки").tag(0)
                        Text("Участники").tag(1)
                        if event.status != "completed" {
                            Text("Завершение").tag(2)
                        }
                    }
                    .pickerStyle(.segmented)
                    .padding()
                    .colorScheme(.dark)
                    
                    if let error = errorMessage {
                        Text(error)
                            .foregroundColor(.red)
                            .font(.system(size: 14))
                            .padding()
                    }
                    
                    TabView(selection: $selectedTab) {
                        editSection.tag(0)
                        playersSection.tag(1)
                        if event.status != "completed" {
                            completeSection.tag(2)
                        }
                    }
                    .tabViewStyle(.page(indexDisplayMode: .never))
                }
            }
            .navigationTitle("Управление")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Готово") { dismiss() }
                }
            }
        }
        .task {
            await loadPlayers()
        }
        .sheet(isPresented: $showAddPlayerSheet) {
            AddPlayerSheet(
                viewModel: viewModel,
                eventId: event.id,
                currentPlayers: $eventPlayers,
                onComplete: {
                    Task { await loadPlayers() }
                }
            )
        }
    }
    
    // MARK: - Sections
    
    private var editSection: some View {
        ScrollView {
            VStack(spacing: 20) {
                VStack(alignment: .leading, spacing: 16) {
                    CustomTextField(placeholder: "Название", text: $title)
                    CustomTextField(placeholder: "Описание", text: $description, isEditor: true)
                    CustomTextField(placeholder: "Базовая награда", text: $rewardString, keyboardType: .numberPad)
                    
                    HStack {
                        Text("Статус")
                            .foregroundStyle(.white.opacity(0.8))
                        Spacer()
                        Picker("Статус", selection: $status) {
                            Text("Ожидается").tag("scheduled")
                            Text("Активно").tag("active")
                            Text("Завершено").tag("completed")
                        }
                        .tint(.blue)
                    }
                    .padding()
                    .background(Color.white.opacity(0.06))
                    .cornerRadius(12)
                    
                    HStack {
                        Text("Начало")
                            .foregroundStyle(.white.opacity(0.8))
                        Spacer()
                        DatePicker("", selection: $startDate, displayedComponents: [.date, .hourAndMinute])
                            .labelsHidden()
                            .colorScheme(.dark)
                    }
                    .padding()
                    .background(Color.white.opacity(0.06))
                    .cornerRadius(12)
                }
                
                Button {
                    saveChanges()
                } label: {
                    Text(isSubmitting ? "Сохранение..." : "Сохранить изменения")
                        .font(.system(size: 16, weight: .bold))
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.blue)
                        .foregroundStyle(.white)
                        .cornerRadius(12)
                }
                .disabled(isSubmitting)
            }
            .padding(20)
        }
    }
    
    private var playersSection: some View {
        VStack {
            if isLoadingPlayers {
                ProgressView().padding()
            } else if eventPlayers.isEmpty {
                Text("Нет участников")
                    .foregroundColor(.white.opacity(0.5))
                    .padding()
            } else {
                List {
                    ForEach(eventPlayers) { player in
                        HStack {
                            VStack(alignment: .leading) {
                                Text(player.fullName.first.map { "\($0.LastName) \($0.Name)" } ?? player.name)
                                    .foregroundStyle(.white)
                                Text(player.login)
                                    .font(.caption)
                                    .foregroundStyle(.gray)
                            }
                            Spacer()
                            Text("\(player.rating) очков")
                                .font(.caption.bold())
                                .foregroundStyle(.blue)
                        }
                        .listRowBackground(Color.white.opacity(0.06))
                        .swipeActions(edge: .trailing) {
                            Button(role: .destructive) {
                                removePlayer(playerId: player.id)
                            } label: {
                                Label("Удалить", systemImage: "trash")
                            }
                        }
                    }
                }
                .scrollContentBackground(.hidden)
            }
            
            Button {
                showAddPlayerSheet = true
            } label: {
                Label("Добавить участников", systemImage: "person.badge.plus")
                    .font(.system(size: 16, weight: .bold))
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.white.opacity(0.1))
                    .foregroundStyle(.blue)
                    .cornerRadius(12)
            }
            .padding(.horizontal, 20)
            .padding(.bottom, 20)
        }
    }
    
    private var completeSection: some View {
        ScrollView {
            VStack(spacing: 20) {
                Text("Завершение мероприятия начислит указанные очки всем текущим участникам (\(eventPlayers.count) чел.) и изменит статус на 'Завершено'.")
                    .font(.system(size: 14))
                    .foregroundStyle(.white.opacity(0.7))
                    .multilineTextAlignment(.center)
                    .padding(.bottom, 10)
                
                VStack(alignment: .leading, spacing: 16) {
                    CustomTextField(placeholder: "Личный рейтинг (каждому)", text: $ratingRewardString, keyboardType: .numberPad)
                    CustomTextField(placeholder: "Рейтинг класса", text: $classRewardString, keyboardType: .numberPad)
                }
                
                Button {
                    completeEvent()
                } label: {
                    Text(isSubmitting ? "Обработка..." : "Завершить и выдать награды")
                        .font(.system(size: 16, weight: .bold))
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.green)
                        .foregroundStyle(.white)
                        .cornerRadius(12)
                }
                .disabled(isSubmitting || eventPlayers.isEmpty)
            }
            .padding(20)
        }
    }
    
    // MARK: - Actions
    
    private func saveChanges() {
        guard let reward = Int(rewardString) else { return }
        isSubmitting = true
        
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        let dateString = formatter.string(from: startDate)
        
        Task {
            do {
                try await viewModel.updateEvent(
                    eventId: event.id,
                    title: title,
                    description: description,
                    reward: reward,
                    status: status,
                    startedAt: dateString,
                    classes: Array(selectedClasses),
                    players: event.players
                )
                errorMessage = "Успешно сохранено"
            } catch {
                errorMessage = "Ошибка сохранения: \(error.localizedDescription)"
            }
            isSubmitting = false
        }
    }
    
    private func loadPlayers() async {
        isLoadingPlayers = true
        
        if viewModel.allUsers.isEmpty {
            await viewModel.fetchAllUsers()
        }
        
        eventPlayers = viewModel.getEventPlayers(for: event)
        
        isLoadingPlayers = false
    }
    
    private func removePlayer(playerId: Int) {
        Task {
            do {
                try await viewModel.removePlayersFromEvent(eventId: event.id, playerIds: [playerId])
                await loadPlayers()
            } catch {
                errorMessage = "Ошибка при удалении участника"
            }
        }
    }
    
    private func completeEvent() {
        guard let rReward = Int(ratingRewardString), let cReward = Int(classRewardString) else {
            errorMessage = "Награды должны быть числами"
            return
        }
        
        isSubmitting = true
        Task {
            do {
                try await viewModel.completeEvent(eventId: event.id, ratingReward: rReward, classReward: cReward)
                dismiss()
            } catch {
                errorMessage = "Ошибка при завершении: \(error.localizedDescription)"
                isSubmitting = false
            }
        }
    }
}
