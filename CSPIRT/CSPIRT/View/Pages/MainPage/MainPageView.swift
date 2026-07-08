import SwiftUI

struct MainPageView: View {
    @StateObject private var viewModel = MainPageViewModel()
    @StateObject private var eventViewModel = EventViewModel()
    @Environment(\.colorScheme) private var colorScheme
    
    @State private var selectedEvent: EventModel?
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }
    
    var body: some View {
        NavigationStack {
            ZStack {
                backgroundLayer
                    .ignoresSafeArea()
                
                if viewModel.isLoading && viewModel.topClassUsers.isEmpty {
                    VStack(spacing: 16) {
                        ProgressView()
                            .scaleEffect(1.5)
                            .tint(colorScheme == .dark ? .white : accentBlue)
                        Text("Загрузка дашборда...")
                            .font(.system(size: 16, weight: .medium, design: .rounded))
                            .foregroundStyle(colorScheme == .dark ? .white.opacity(0.6) : .black.opacity(0.6))
                    }
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    
                } else if let errMsg = viewModel.errMsg, viewModel.topClassUsers.isEmpty {
                    VStack(spacing: 20) {
                        Image(systemName: "wifi.exclamationmark")
                            .font(.system(size: 40, weight: .medium))
                            .foregroundStyle(colorScheme == .dark ? .white.opacity(0.7) : .black.opacity(0.6))
                        
                        Text(errMsg)
                            .font(.system(size: 16, weight: .medium, design: .rounded))
                            .foregroundStyle(colorScheme == .dark ? .white.opacity(0.9) : .black.opacity(0.9))
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 32)
                        
                        Button {
                            Task { await viewModel.loadDashboardData() }
                        } label: {
                            Text("Повторить попытку")
                                .font(.system(size: 15, weight: .semibold, design: .rounded))
                                .foregroundStyle(.white) // Оставляем белым на цветной кнопке
                                .padding(.horizontal, 24)
                                .padding(.vertical, 12)
                                .background(accentBlue)
                                .cornerRadius(12)
                        }
                        .shadow(color: accentBlue.opacity(0.3), radius: 8, x: 0, y: 4)
                    }
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    
                } else {
                    ScrollView(.vertical, showsIndicators: false) {
                        VStack(spacing: 16) {
                            headerView
                                .padding(.horizontal, 15)
                                .padding(.top, 10)
                                .padding(.bottom, 10)
                            
                            summaryStrip
                                .padding(.horizontal, 15)
                            
                            if !viewModel.topClassUsers.isEmpty {
                                leaderboardSection
                                    .padding(.top, 5)
                                    .padding(.horizontal, 15)
                            }
                            
                            eventsSection
                                .padding(.bottom, 5)
                                .padding(.top, 10)
                                .padding(.horizontal, 15)
                            
                            disciplineSection
                                .padding(.horizontal, 15)
                        }
                        .padding(.horizontal, 16)
                        .padding(.top, 16)
                        .padding(.bottom, 24)
                    }
//                        .safeAreaInset(edge: .bottom) {
//                            Color.clear.frame(height: 90)
//                        }
                    .refreshable {
                        await viewModel.loadDashboardData()
                    }
                }
            }
            .task {
                await viewModel.loadDashboardData()
            }
            .sheet(item: $selectedEvent) { eventData in
                EventDetailView(event: eventData, allClasses: eventViewModel.classesInEvents)
            }
        }
    }
}

// MARK: - UI Components

private extension MainPageView {
    
    var backgroundLayer: some View {
        ZStack {
            Image("school_background")
                .resizable()
                .scaledToFill()
                .scaleEffect(colorScheme == .dark ? 1.08 : 1.0)
                .blur(radius: colorScheme == .dark ? 18 : 16)
                .overlay {
                    if colorScheme == .dark {
                        LinearGradient(
                            colors: [
                                .black.opacity(0.35),
                                .black.opacity(0.70),
                                .black.opacity(0.92)
                            ],
                            startPoint: .top,
                            endPoint: .bottom
                        )
                    } else {
                        Color.white.opacity(0.65)
                    }
                }

            Color.black.opacity(colorScheme == .dark ? 0.12 : 0.0)
        }
    }

    var headerView: some View {
        GlassCard {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Главная")
                        .font(.system(size: 28, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

                    Text("Короткая сводка по классу, событиям и дисциплине")
                        .font(.system(size: 14, weight: .regular))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.62) : .black.opacity(0.6))
                        .fixedSize(horizontal: false, vertical: true)
                }

                Spacer(minLength: 0)

                Image(systemName: "house.fill")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundStyle(accentBlue)
                    .frame(width: 40, height: 40)
                    .background(colorScheme == .dark ? Color.white.opacity(0.06) : Color.black.opacity(0.05))
                    .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            }
        }
    }

    var summaryStrip: some View {
        HStack(spacing: 12) {
            var ClassModel = MyClassViewModel()
            statCard(title: "Рейтинг", value: "\(ClassModel.userTotalRating)", icon: "chart.bar.fill")
            statCard(title: "Мероприятий", value: "\(viewModel.availableEvents.count)", icon: "calendar.badge.clock")
        }
    }

    func statCard(title: String, value: String, icon: String) -> some View {
        GlassCard(padding: 14) {
            VStack(alignment: .leading, spacing: 10) {
                Image(systemName: icon)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(accentBlue)
                    .frame(width: 30, height: 30)
                    .background(colorScheme == .dark ? Color.white.opacity(0.06) : Color.black.opacity(0.05))
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))

                VStack(alignment: .leading, spacing: 2) {
                    Text(value)
                        .font(.system(size: 20, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

                    Text(title)
                        .font(.system(size: 12, weight: .medium))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.52) : .black.opacity(0.5))
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    var eventsSection: some View {
        SectionCard(title: "Предстоящие мероприятия", icon: "calendar") {
            if viewModel.availableEvents.isEmpty {
                EmptyStateView(
                    icon: "calendar.badge.exclamationmark",
                    title: "Пока ничего нет",
                    subtitle: "Когда появятся мероприятия, они будут отображаться здесь."
                )
            } else {
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 12) {
                        ForEach(viewModel.availableEvents, id: \.id) { event in
                            EventCard(event: event, onButtonTap: {
                                self.selectedEvent = event
                            })
                        }
                    }
                    .padding(.vertical, 2)
                }
            }
        }
    }

    var leaderboardSection: some View {
        SectionCard(title: "Топ-5 класса", icon: "trophy.fill") {
            if viewModel.topClassUsers.isEmpty {
                EmptyStateView(
                    icon: "person.3.fill",
                    title: "Пока нет данных",
                    subtitle: "Рейтинг появится после загрузки информации о классе."
                )
            } else {
                VStack(spacing: 10) {
                    ForEach(Array(viewModel.topClassUsers.enumerated()), id: \.offset) { index, user in
                        LeaderboardRow(
                            place: index + 1,
                            name: "\(user.lastName) \(user.name)",
                            rating: user.rating,
                            accent: rankColor(for: index)
                        )

                        if index < viewModel.topClassUsers.count - 1 {
                            Divider()
                                .overlay(colorScheme == .dark ? Color.white.opacity(0.08) : Color.black.opacity(0.05))
                        }
                    }
                }
            }
        }
    }

    var disciplineSection: some View {
        SectionCard(title: "Дисциплина", icon: "shield.lefthalf.filled") {
            VStack(spacing: 16) {
                notesBlock
                complaintsBlock
            }
        }
    }

    var notesBlock: some View {
        VStack(alignment: .leading, spacing: 10) {
            SectionLabel(
                text: "Последние заметки",
                systemImage: "bookmark.fill",
                tint: .green
            )

            if viewModel.latestNotes.isEmpty {
                EmptyInlineText(text: "Заметок пока нет")
            } else {
                VStack(spacing: 8) {
                    ForEach(viewModel.latestNotes, id: \.id) { note in
                        ItemRow(
                            accent: .green,
                            text: note.content
                        )
                    }
                }
            }
        }
    }

    var complaintsBlock: some View {
        VStack(alignment: .leading, spacing: 10) {
            SectionLabel(
                text: "Последние жалобы",
                systemImage: "exclamationmark.triangle.fill",
                tint: .red
            )

            if viewModel.latestComplaints.isEmpty {
                EmptyInlineText(text: "Жалоб нет. Ситуация под контролем.")
            } else {
                VStack(spacing: 8) {
                    ForEach(viewModel.latestComplaints, id: \.id) { complaint in
                        ItemRow(
                            accent: .red,
                            text: complaint.content
                        )
                    }
                }
            }
        }
    }

    func rankColor(for index: Int) -> Color {
        switch index {
        case 0: return .yellow
        case 1: return .gray
        case 2: return .orange
        default: return colorScheme == .dark ? Color.white.opacity(0.55) : Color.black.opacity(0.4)
        }
    }
}

// MARK: - Reusable Views

private struct GlassCard<Content: View>: View {
    var padding: CGFloat = 16
    @Environment(\.colorScheme) private var colorScheme
    @ViewBuilder let content: Content

    var body: some View {
        content
            .padding(padding)
            .background(
                RoundedRectangle(cornerRadius: 22, style: .continuous)
                    .fill(colorScheme == .dark
                          ? Color.white.opacity(0.07)
                          : Color.white.opacity(0.6))
                    .overlay(
                        RoundedRectangle(cornerRadius: 22, style: .continuous)
                            .stroke(colorScheme == .dark
                                    ? Color.white.opacity(0.10)
                                    : Color.black.opacity(0.06), lineWidth: 1)
                    )
                    .shadow(color: .black.opacity(colorScheme == .dark ? 0.18 : 0.06),
                            radius: 18, x: 0, y: 10)
            )
    }
}

private struct SectionCard<Content: View>: View {
    let title: String
    let icon: String
    @Environment(\.colorScheme) private var colorScheme
    @ViewBuilder let content: Content
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }

    var body: some View {
        GlassCard {
            VStack(alignment: .leading, spacing: 14) {
                HStack(spacing: 10) {
                    Image(systemName: icon)
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(accentBlue)
                        .frame(width: 28, height: 28)
                        .background(colorScheme == .dark ? Color.white.opacity(0.06) : Color.black.opacity(0.05))
                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))

                    Text(title)
                        .font(.system(size: 17, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

                    Spacer(minLength: 0)
                }

                content
            }
        }
    }
}

private struct SectionLabel: View {
    let text: String
    let systemImage: String
    let tint: Color
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: systemImage)
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(tint.opacity(0.9))

            Text(text)
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

            Spacer(minLength: 0)
        }
    }
}

private struct EmptyStateView: View {
    let icon: String
    let title: String
    let subtitle: String
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: icon)
                .font(.system(size: 20, weight: .semibold))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.7) : .black.opacity(0.5))
                .frame(width: 44, height: 44)
                .background(colorScheme == .dark ? Color.white.opacity(0.05) : Color.black.opacity(0.04))
                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))

            Text(title)
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

            Text(subtitle)
                .font(.system(size: 13))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.5) : .black.opacity(0.5))
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 12)
    }
}

private struct EmptyInlineText: View {
    let text: String
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        Text(text)
            .font(.system(size: 13))
            .foregroundStyle(colorScheme == .dark ? .white.opacity(0.5) : .black.opacity(0.5))
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.vertical, 4)
    }
}

private struct ItemRow: View {
    let accent: Color
    let text: String
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        HStack(alignment: .top, spacing: 10) {
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .fill(accent)
                .frame(width: 4, height: 34)

            Text(text)
                .font(.system(size: 14))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.88) : .black.opacity(0.8))
                .frame(maxWidth: .infinity, alignment: .leading)

            Spacer(minLength: 0)
        }
        .padding(12)
        .background(colorScheme == .dark ? Color.white.opacity(0.04) : Color.black.opacity(0.03))
        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
    }
}

private struct LeaderboardRow: View {
    let place: Int
    let name: String
    let rating: Int
    let accent: Color
    @Environment(\.colorScheme) private var colorScheme
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }

    var body: some View {
        HStack(spacing: 12) {
            Text("\(place)")
                .font(.system(size: 13, weight: .bold, design: .rounded))
                .foregroundStyle(accent)
                .frame(width: 28, height: 28)
                .background(colorScheme == .dark ? Color.white.opacity(0.05) : Color.black.opacity(0.04))
                .clipShape(Circle())

            Text(name)
                .font(.system(size: 15, weight: .medium))
                .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                .lineLimit(1)

            Spacer(minLength: 0)

            Text("\(rating)")
                .font(.system(size: 15, weight: .bold, design: .rounded))
                .foregroundStyle(accentBlue)
        }
        .padding(.vertical, 2)
    }
}

private struct EventCard: View {
    let event: EventModel
    let onButtonTap: () -> Void
    
    @Environment(\.colorScheme) private var colorScheme
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }
    
    var body: some View {
        GlassCard(padding: 14) {
            VStack(alignment: .leading, spacing: 12) {
                HStack(alignment: .top, spacing: 8) {
                    Text(event.title)
                        .font(.system(size: 16, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                        .lineLimit(2)

                    Spacer(minLength: 0)

                    Text("+\(event.baseRatingReward)")
                        .font(.system(size: 11, weight: .bold))
                        .foregroundStyle(accentBlue)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(accentBlue.opacity(0.15))
                        .clipShape(Capsule())
                }

                Text(event.description)
                    .font(.system(size: 13))
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.64) : .black.opacity(0.6))
                    .lineLimit(3)
                    .fixedSize(horizontal: false, vertical: true)

                Button {
                    onButtonTap()
                } label: {
                    Text("Подробнее")
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .frame(height: 38)
                        .background(accentBlue)
                        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
                }
                .buttonStyle(.plain)
            }
            .frame(width: 246, alignment: .leading)
        }
    }
}
