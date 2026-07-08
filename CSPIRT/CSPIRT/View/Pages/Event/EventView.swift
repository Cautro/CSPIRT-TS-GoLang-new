// EventView.swift
import SwiftUI

struct EventView: View {
    @StateObject private var viewModel = EventViewModel()
    
    // Состояния навигации и модалок
    @State private var selectedEvent: EventModel?
    @State private var eventToManage: EventModel?
    @State private var eventToDelete: EventModel?
    
    @State private var showCreateSheet = false
    @State private var showDeleteConfirmation = false
    
    private var completedPastEvents: [EventModel] {
        viewModel.pastEvents.filter { $0.status == "completed" }
    }

    var body: some View {
        ZStack {
            backgroundLayer
                .ignoresSafeArea()
            
            if viewModel.isLoading && viewModel.availableEvents.isEmpty && viewModel.pastEvents.isEmpty {
                loadingView
            } else if let errMsg = viewModel.errMsg, viewModel.availableEvents.isEmpty, viewModel.pastEvents.isEmpty {
                VStack(spacing: 20) {
                    Image(systemName: "wifi.exclamationmark")
                        .font(.system(size: 40, weight: .medium))
                        .foregroundStyle(.white.opacity(0.7))
                    
                    Text(errMsg)
                        .font(.system(size: 16, weight: .medium, design: .rounded))
                        .foregroundStyle(.white.opacity(0.9))
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 32)
                    
                    Button {
                        Task { await viewModel.fetchEvents() }
                    } label: {
                        Text("Повторить попытку")
                            .font(.system(size: 15, weight: .semibold, design: .rounded))
                            .foregroundStyle(.white)
                            .padding(.horizontal, 24)
                            .padding(.vertical, 12)
                            .background(Color(red: 0.0, green: 0.55, blue: 1.0))
                            .cornerRadius(12)
                    }
                    .shadow(color: Color(red: 0.0, green: 0.55, blue: 1.0).opacity(0.3), radius: 8, x: 0, y: 4)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollView(.vertical, showsIndicators: false) {
                    VStack(spacing: 20) {
                        headerView
                            .padding(.horizontal, 15)
                            .padding(.top, 10)
                        
                        if let featured = viewModel.featuredEvent {
                            featuredEventSection(event: featured)
                                .padding(.horizontal, 15)
                        }
                        
                        allEventsSection
                            .padding(.horizontal, 15)
                        
                        if !completedPastEvents.isEmpty {
                            pastEventsSection
                                .padding(.horizontal, 15)
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.top, 16)
                    .padding(.bottom, 32)
                }
                .refreshable {
                    await viewModel.fetchEvents()
                }
            }
        }
        .task {
            await viewModel.fetchEvents()
        }
        .sheet(item: $selectedEvent) { event in
            EventDetailView(event: event, allClasses: viewModel.classesInEvents)
        }
        .sheet(isPresented: $showCreateSheet) {
            CreateEventSheet(viewModel: viewModel)
        }
        .sheet(item: $eventToManage) { event in
            ManageEventSheet(viewModel: viewModel, event: event)
        }
        .confirmationDialog(
            "Удалить мероприятие?",
            isPresented: $showDeleteConfirmation,
            titleVisibility: .visible
        ) {
            Button("Удалить", role: .destructive) {
                if let eventId = eventToDelete?.id {
                    Task { try await viewModel.deleteEvent(eventId: eventId) }
                }
            }
            Button("Отмена", role: .cancel) {
                eventToDelete = nil
            }
        } message: {
            Text("Это действие нельзя отменить. Все данные об участниках будут стерты.")
        }
        .onChange(of: eventToDelete) { newValue in
            if newValue != nil {
                showDeleteConfirmation = true
            }
        }
    }
}

// MARK: - UI Слой

private extension EventView {
    var backgroundLayer: some View {
        ZStack {
            Image("school_background")
                .resizable()
                .scaledToFill()
                .scaleEffect(1.08)
                .blur(radius: 18)
                .overlay {
                    LinearGradient(
                        colors: [
                            Color.black.opacity(0.4),
                            Color.black.opacity(0.75),
                            Color.black.opacity(0.95)
                        ],
                        startPoint: .top,
                        endPoint: .bottom
                    )
                }
            Color.black.opacity(0.15)
        }
    }
    
    var loadingView: some View {
        VStack(spacing: 16) {
            ProgressView()
                .scaleEffect(1.5)
                .tint(.white)
            Text("Загрузка событий...")
                .font(.system(size: 16, weight: .medium, design: .rounded))
                .foregroundStyle(.white.opacity(0.6))
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    var headerView: some View {
        GlassEventCard {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 6) {
                    Text("Мероприятия")
                        .font(.system(size: 28, weight: .bold, design: .rounded))
                        .foregroundStyle(.white)

                    Text("Участвуйте в событиях и зарабатывайте очки рейтинга")
                        .font(.system(size: 14, weight: .regular))
                        .foregroundStyle(.white.opacity(0.6))
                        .fixedSize(horizontal: false, vertical: true)
                }

                Spacer(minLength: 4)

                if viewModel.isOwner {
                    Menu {
                        Button {
                            showCreateSheet = true
                        } label: {
                            Label("Создать мероприятие", systemImage: "plus.circle")
                        }
                        
                        Button {
                            Task { await viewModel.fetchEvents() }
                        } label: {
                            Label("Обновить список", systemImage: "arrow.triangle.2.circlepath")
                        }
                    } label: {
                        Image(systemName: "slider.horizontal.3")
                            .font(.system(size: 18, weight: .semibold))
                            .foregroundStyle(.blue)
                            .frame(width: 42, height: 42)
                            .background(Color.white.opacity(0.06))
                            .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                    }
                } else {
                    Image(systemName: "trophy.fill")
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundStyle(.blue)
                        .frame(width: 42, height: 42)
                        .background(Color.white.opacity(0.06))
                        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                }
            }
        }
    }
    
    func featuredEventSection(event: EventModel) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("АКТУАЛЬНОЕ СОБЫТИЕ")
                .font(.system(size: 12, weight: .bold, design: .rounded))
                .foregroundStyle(.blue)
                .tracking(1.5)
                .padding(.leading, 4)
            
            GlassEventCard(padding: 20.0) {
                VStack(alignment: .leading, spacing: 16) {
                    HStack(alignment: .top) {
                        VStack(alignment: .leading, spacing: 4) {
                            Text(event.title)
                                .font(.system(size: 22, weight: .bold, design: .rounded))
                                .foregroundStyle(.white)
                            
                            HStack(spacing: 6) {
                                Image(systemName: "clock")
                                Text("Начало: \(formatDate(event.startedAt))")
                            }
                            .font(.system(size: 12, weight: .medium))
                            .foregroundStyle(.white.opacity(0.5))
                        }
                        
                        Spacer()
                        
                        HStack(spacing: 4) {
                            Text("+\(event.baseRatingReward)")
                                .font(.system(size: 16, weight: .heavy, design: .rounded))
                            Image(systemName: "star.fill")
                                .font(.system(size: 12))
                        }
                        .foregroundStyle(.blue)
                        .padding(.horizontal, 12)
                        .padding(.vertical, 6)
                        .background(Color.blue.opacity(0.15))
                        .cornerRadius(10)
                    }
                    
                    Text(event.description)
                        .font(.system(size: 14))
                        .foregroundStyle(.white.opacity(0.8))
                        .lineSpacing(4)
                    
                    Divider()
                        .background(Color.white.opacity(0.1))
                    
                    HStack {
                        Label("\(event.players.count) участников", systemImage: "person.3.fill")
                        Spacer()
                        
                        Text(event.statusDisplayName)
                            .font(.system(size: 11, weight: .bold, design: .rounded))
                            .padding(.horizontal, 10)
                            .padding(.vertical, 4)
                            .background(statusBackgroundColor(for: event.status))
                            .foregroundStyle(statusForegroundColor(for: event.status))
                            .cornerRadius(8)
                    }
                    .font(.system(size: 13, weight: .medium))
                    .foregroundStyle(.white.opacity(0.6))
                }
            }
            .overlay(
                RoundedRectangle(cornerRadius: 22)
                    .stroke(LinearGradient(colors: [.blue.opacity(0.4), .clear], startPoint: .topLeading, endPoint: .bottomTrailing), lineWidth: 1.5)
            )
            .onTapGesture {
                self.selectedEvent = event
            }
            .contextMenu {
                Button {
                    self.selectedEvent = event
                } label: {
                    Label("Подробнее", systemImage: "info.circle")
                }
                
                if viewModel.isOwner {
                    Button {
                        self.eventToManage = event
                    } label: {
                        Label("Управление", systemImage: "slider.horizontal.3")
                    }
                    
                    Button(role: .destructive) {
                        self.eventToDelete = event
                    } label: {
                        Label("Удалить", systemImage: "trash")
                    }
                }
            }
        }
    }
    
    var allEventsSection: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("ВСЕ ДОСТУПНЫЕ СОБЫТИЯ")
                .font(.system(size: 12, weight: .bold, design: .rounded))
                .foregroundStyle(.white.opacity(0.4))
                .tracking(1.5)
                .padding(.leading, 4)
            
            if viewModel.availableEvents.count <= 1 {
                GlassEventCard {
                    HStack {
                        Spacer()
                        Text("Других событий пока нет")
                            .font(.system(size: 14, weight: .medium))
                            .foregroundStyle(.white.opacity(0.4))
                            .padding(.vertical, 10)
                        Spacer()
                    }
                }
            } else {
                ForEach(viewModel.availableEvents.dropFirst(), id: \.id) { event in
                    EventRowCard(event: event)
                        .onTapGesture {
                            self.selectedEvent = event
                        }
                        .contextMenu {
                            Button {
                                self.selectedEvent = event
                            } label: {
                                Label("Подробнее", systemImage: "info.circle")
                            }
                            
                            if viewModel.isOwner {
                                Button {
                                    self.eventToManage = event
                                } label: {
                                    Label("Управление", systemImage: "slider.horizontal.3")
                                }
                                
                                Button(role: .destructive) {
                                    self.eventToDelete = event
                                } label: {
                                    Label("Удалить", systemImage: "trash")
                                }
                            }
                        }
                }
            }
        }
    }
    
    var pastEventsSection: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("ПРОШЕДШИЕ СОБЫТИЯ")
                .font(.system(size: 12, weight: .bold, design: .rounded))
                .foregroundStyle(.white.opacity(0.4))
                .tracking(1.5)
                .padding(.leading, 4)
                .padding(.top, 10)
            
            ForEach(completedPastEvents, id: \.id) { event in
                EventRowCard(event: event)
                    .opacity(0.6)
                    .grayscale(0.8)
                    .onTapGesture {
                        self.selectedEvent = event
                    }
                    .contextMenu {
                        Button {
                            self.selectedEvent = event
                        } label: {
                            Label("Подробнее", systemImage: "info.circle")
                        }
                        
                        if viewModel.isOwner {
                            Button {
                                self.eventToManage = event
                            } label: {
                                Label("Управление", systemImage: "slider.horizontal.3")
                            }
                            
                            Button(role: .destructive) {
                                self.eventToDelete = event
                            } label: {
                                Label("Удалить", systemImage: "trash")
                            }
                        }
                    }
            }
        }
    }
    
    func formatDate(_ isoString: String) -> String {
        guard !isoString.isEmpty else { return "--.--.----" }
        return String(isoString.prefix(10)).replacingOccurrences(of: "-", with: ".")
    }
    
    // Вспомогательные функции для цветов UI
    func statusBackgroundColor(for status: String) -> Color {
        switch status {
        case "active": return .green.opacity(0.15)
        case "scheduled": return .orange.opacity(0.15)
        case "completed": return .gray.opacity(0.2)
        default: return .white.opacity(0.1)
        }
    }
    
    func statusForegroundColor(for status: String) -> Color {
        switch status {
        case "active": return .green
        case "scheduled": return .orange
        case "completed": return .gray
        default: return .white
        }
    }
}

// MARK: - Карточка строки события

private struct EventRowCard: View {
    let event: EventModel
    
    var body: some View {
        GlassEventCard(padding: 0.0) {
            HStack(alignment: .center, spacing: 0) {
                Rectangle()
                    .fill(event.isActive ? Color.green : Color.white.opacity(0.3))
                    .frame(width: 4)
                
                VStack(alignment: .leading, spacing: 10) {
                    HStack(alignment: .top) {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(event.title)
                                .font(.system(size: 16, weight: .semibold, design: .rounded))
                                .foregroundStyle(.white)
                            
                            Text("Классы: \(event.classes.isEmpty ? "Все" : event.classes.map { String($0) }.joined(separator: ", "))")
                                .font(.system(size: 12))
                                .foregroundStyle(.white.opacity(0.45))
                        }
                        
                        Spacer()
                        
                        HStack(spacing: 3) {
                            Text("+\(event.baseRatingReward)")
                                .font(.system(size: 14, weight: .bold, design: .rounded))
                            Image(systemName: "star.fill")
                                .font(.system(size: 10))
                        }
                        .foregroundStyle(.blue.opacity(0.9))
                        .padding(.horizontal, 10)
                        .padding(.vertical, 4)
                        .background(Color.white.opacity(0.05))
                        .cornerRadius(8)
                    }
                    
                    HStack {
                        Label("\(event.players.count) уч.", systemImage: "person.2.fill")
                        Spacer()
                        
                        Text(String(event.startedAt.prefix(10)).replacingOccurrences(of: "-", with: "."))
                    }
                    .font(.system(size: 12, weight: .regular))
                    .foregroundStyle(.white.opacity(0.5))
                }
                .padding(16)
            }
        }
        .contentShape(Rectangle())
        .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
    }
}

// MARK: - Компонент Glassmorphism-карточки

struct GlassEventCard<Content: View>: View {
    let padding: CGFloat
    let content: Content

    init(padding: CGFloat = 16.0, @ViewBuilder content: () -> Content) {
        self.padding = padding
        self.content = content()
    }

    var body: some View {
        content
            .padding(padding)
            .background(
                RoundedRectangle(cornerRadius: 22, style: .continuous)
                    .fill(Color.white.opacity(0.06))
                    .overlay(
                        RoundedRectangle(cornerRadius: 22, style: .continuous)
                            .stroke(Color.white.opacity(0.08), lineWidth: 1)
                    )
                    .shadow(color: .black.opacity(0.2), radius: 15, x: 0, y: 8)
            )
    }
}

struct EventDetailView: View {
    let event: EventModel
    let allClasses: [ClassesModel]
    @Environment(\.dismiss) private var dismiss
    
    var body: some View {
        ZStack {
            Color(red: 0.05, green: 0.05, blue: 0.07)
                .ignoresSafeArea()
            
            Circle()
                .fill(Color.blue.opacity(0.12))
                .frame(width: 320, height: 320)
                .blur(radius: 70)
                .offset(x: -80, y: -150)
            
            VStack(spacing: 0) {
                HStack {
                    Text("Детали события")
                        .font(.system(size: 17, weight: .bold, design: .rounded))
                        .foregroundStyle(.white)
                    Spacer()
                    Button {
                        dismiss()
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .font(.system(size: 24))
                            .foregroundStyle(.white.opacity(0.4))
                    }
                }
                .padding(.horizontal, 20)
                .padding(.top, 20)
                .padding(.bottom, 10)
                
                ScrollView(.vertical, showsIndicators: false) {
                    VStack(alignment: .leading, spacing: 24) {
                        
                        GlassEventCard(padding: 20) {
                            VStack(alignment: .leading, spacing: 12) {
                                HStack(alignment: .top) {
                                    Text(event.title)
                                        .font(.system(size: 24, weight: .bold, design: .rounded))
                                        .foregroundStyle(.white)
                                        .fixedSize(horizontal: false, vertical: true)
                                    
                                    Spacer(minLength: 10)
                                    
                                    HStack(spacing: 4) {
                                        Text("+\(event.baseRatingReward)")
                                            .font(.system(size: 18, weight: .heavy, design: .rounded))
                                        Image(systemName: "star.fill")
                                            .font(.system(size: 14))
                                    }
                                    .foregroundStyle(.blue)
                                    .padding(.horizontal, 12)
                                    .padding(.vertical, 6)
                                    .background(Color.blue.opacity(0.15))
                                    .cornerRadius(10)
                                }
                                
                                HStack(spacing: 12) {
                                    Label(event.statusDisplayName, systemImage: "info.circle")
                                        .font(.system(size: 12, weight: .bold))
                                        .padding(.horizontal, 8)
                                        .padding(.vertical, 4)
                                        .background(statusBackgroundColor(for: event.status))
                                        .foregroundStyle(statusForegroundColor(for: event.status))
                                        .cornerRadius(6)
                                    
                                    Spacer()
                                }
                            }
                        }
                        
                        VStack(alignment: .leading, spacing: 10) {
                            Text("ОПИСАНИЕ")
                                .font(.system(size: 12, weight: .bold, design: .rounded))
                                .foregroundStyle(.white.opacity(0.4))
                                .tracking(1.5)
                                .padding(.leading, 4)
                            
                            GlassEventCard(padding: 20) {
                                Text(event.description.isEmpty ? "Описание для данного события отсутствует." : event.description)
                                    .font(.system(size: 15))
                                    .foregroundStyle(.white.opacity(0.85))
                                    .lineSpacing(6)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                            }
                        }
                        
                        VStack(alignment: .leading, spacing: 10) {
                            Text("ИНФОРМАЦИЯ")
                                .font(.system(size: 12, weight: .bold, design: .rounded))
                                .foregroundStyle(.white.opacity(0.4))
                                .tracking(1.5)
                                .padding(.leading, 4)
                            
                            GlassEventCard(padding: 16) {
                                VStack(spacing: 14) {
                                    infoRow(icon: "calendar", title: "Дата старта", value: cleanDate(event.startedAt))
                                    Divider().background(Color.white.opacity(0.06))
                                    
                                    infoRow(icon: "person.2.fill", title: "Участники", value: "\(event.players.count) чел.")
                                    Divider().background(Color.white.opacity(0.06))
                                    
                                    infoRow(
                                        icon: "graduationcap.fill",
                                        title: "Доступно классам",
                                        value: {
                                            if allClasses.isEmpty && !event.classes.isEmpty {
                                                return "Загрузка..."
                                            }
                                            
                                            let matchedClasses = allClasses.filter { event.classes.contains($0.id) }
                                            
                                            if matchedClasses.isEmpty {
                                                return "Всем классам"
                                            }
                                            
                                            return matchedClasses
                                                .map { "\($0.grade)\($0.letter)" }
                                                .joined(separator: ", ")
                                        }()
                                    )
                                }
                            }
                        }
                    }
                    .padding(20)
                }
            }
        }
    }
    
    private func infoRow(icon: String, title: String, value: String) -> some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .font(.system(size: 14, weight: .medium))
                .foregroundStyle(.blue.opacity(0.8))
                .frame(width: 28, height: 28)
                .background(Color.white.opacity(0.05))
                .clipShape(RoundedRectangle(cornerRadius: 8))
            
            Text(title)
                .font(.system(size: 14, weight: .regular))
                .foregroundStyle(.white.opacity(0.5))
            
            Spacer()
            
            Text(value)
                .font(.system(size: 14, weight: .semibold, design: .rounded))
                .foregroundStyle(.white.opacity(0.9))
        }
    }
    
    private func cleanDate(_ isoString: String) -> String {
        guard !isoString.isEmpty else { return "--.--.----" }
        return String(isoString.prefix(10)).replacingOccurrences(of: "-", with: ".")
    }
    
    func statusBackgroundColor(for status: String) -> Color {
        switch status {
        case "active": return .green.opacity(0.15)
        case "scheduled": return .orange.opacity(0.15)
        case "completed": return .gray.opacity(0.2)
        default: return .white.opacity(0.1)
        }
    }
    
    func statusForegroundColor(for status: String) -> Color {
        switch status {
        case "active": return .green
        case "scheduled": return .orange
        case "completed": return .gray
        default: return .white
        }
    }
}
