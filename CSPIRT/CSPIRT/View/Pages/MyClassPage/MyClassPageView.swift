import SwiftUI
import Combine

struct MyClassView: View {
    @StateObject private var viewModel = MyClassViewModel()
    @Environment(\.colorScheme) private var colorScheme
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }

    var body: some View {
        ZStack {
            backgroundLayer
                .ignoresSafeArea()
            
            if viewModel.isLoading && viewModel.members.isEmpty {
                loadingView
            } else if let errMsg = viewModel.errMsg, viewModel.members.isEmpty {
                errorView(msg: errMsg)
            } else {
                ScrollView(.vertical, showsIndicators: false) {
                    VStack(spacing: 16) {
                        headerView
                            .padding(.horizontal, 15)
                            .padding(.top, 10)
                            .padding(.bottom, 10)

                        summaryStrip
                            .padding(.bottom, 5)
                            .padding(.horizontal, 15)
                        
                        if viewModel.teacher != nil {
                            teacherSection
                                .padding(.horizontal, 15)
                        }

                        classmatesSection
                            .padding(.horizontal, 15)
                    }
                    .padding(.horizontal, 16)
                    .padding(.top, 16)
                    .padding(.bottom, 24)
                }
                .refreshable {
                    await viewModel.fetchMyClass()
                }
            }
        }
        .task {
            await viewModel.fetchMyClass()
        }
    }
}

// MARK: - UI Components
private extension MyClassView {
    
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
                                Color.black.opacity(0.35),
                                Color.black.opacity(0.70),
                                Color.black.opacity(0.92)
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
    
    var loadingView: some View {
        VStack(spacing: 16) {
            ProgressView()
                .scaleEffect(1.5)
                .tint(colorScheme == .dark ? .white : accentBlue)
            Text("Загрузка класса...")
                .font(.system(size: 16, weight: .medium, design: .rounded))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.6) : .black.opacity(0.6))
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
    
    func errorView(msg: String) -> some View {
        VStack(spacing: 20) {
            Image(systemName: "wifi.exclamationmark")
                .font(.system(size: 40, weight: .medium))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.7) : .black.opacity(0.5))
            
            Text(msg)
                .font(.system(size: 16, weight: .medium, design: .rounded))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.9) : .black.opacity(0.8))
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)
            
            Button {
                Task { await viewModel.fetchMyClass() }
            } label: {
                Text("Повторить попытку")
                    .font(.system(size: 15, weight: .semibold, design: .rounded))
                    .foregroundStyle(.white)
                    .padding(.horizontal, 24)
                    .padding(.vertical, 12)
                    .background(accentBlue)
                    .cornerRadius(12)
            }
            .shadow(color: accentBlue.opacity(0.3), radius: 8, x: 0, y: 4)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    var headerView: some View {
        GlassCard {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Мой класс")
                        .font(.system(size: 28, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

                    Text("\(viewModel.grade)\(viewModel.letter)")
                        .font(.system(size: 14, weight: .regular))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.62) : .black.opacity(0.6))
                        .fixedSize(horizontal: false, vertical: true)
                }

                Spacer(minLength: 0)

                Image(systemName: "graduationcap.fill")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundStyle(accentBlue)
                    .frame(width: 40, height: 40)
                    .background(colorScheme == .dark ? accentBlue.opacity(0.18) : accentBlue.opacity(0.12))
                    .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            }
        }
    }
    
    var summaryStrip: some View {
        HStack(spacing: 12) {
            statCard(title: "Рейтинг класса", value: "\(viewModel.userTotalRating)", icon: "star.fill", color: .yellow)
            statCard(title: "Учеников", value: "\(viewModel.members.count)", icon: "person.2.fill", color: .green)
        }
    }

    func statCard(title: String, value: String, icon: String, color: Color) -> some View {
        GlassCard(padding: 14) {
            VStack(alignment: .leading, spacing: 10) {
                Image(systemName: icon)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(color)
                    .frame(width: 30, height: 30)
                    .background(colorScheme == .dark ? Color.white.opacity(0.06) : color.opacity(0.15))
                    .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))

                VStack(alignment: .leading, spacing: 2) {
                    Text(value)
                        .font(.system(size: 20, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

                    Text(title)
                        .font(.system(size: 11, weight: .medium))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.52) : .black.opacity(0.5))
                        .lineLimit(1)
                        .minimumScaleFactor(0.8)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }
    
    var teacherSection: some View {
        GlassCard(padding: 16) {
            HStack(spacing: 14) {
                Image(systemName: "person.crop.circle.badge.checkmark")
                    .font(.system(size: 32, weight: .light))
                    .foregroundStyle(accentBlue)
                
                VStack(alignment: .leading, spacing: 4) {
                    Text("Классный руководитель")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.52) : .black.opacity(0.5))
                    
                    if let teacher = viewModel.teacher {
                        Text("\(teacher.lastName) \(teacher.name)")
                            .font(.system(size: 16, weight: .semibold, design: .rounded))
                            .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                    }
                }
                Spacer()
            }
        }
    }

    var classmatesSection: some View {
        SectionCard(title: "Одноклассники", icon: "person.3.fill") {
            let sortedMembers = viewModel.members
                .filter { $0.role != UserRole.admin && $0.role != UserRole.owner }
                .sorted { $0.rating > $1.rating }
            
            if sortedMembers.isEmpty {
                EmptyStateView(
                    icon: "person.slash.fill",
                    title: "Пока никого нет",
                    subtitle: "Список учеников появится после загрузки."
                )
            } else {
                VStack(spacing: 10) {
                    ForEach(Array(sortedMembers.enumerated()), id: \.element.id) { index, user in
                        NavigationLink(destination: UserDetailView(userModel: user)) {
                            LeaderboardRow(
                                place: index + 1,
                                name: "\(user.lastName) \(user.name)",
                                rating: user.rating,
                                accent: rankColor(for: index),
                                showChevron: false // Стрелка отключена, как было в оригинале
                            )
                            
                            if index < sortedMembers.count - 1 {
                                Divider()
                                    .overlay(colorScheme == .dark ? Color.white.opacity(0.15) : Color.black.opacity(0.1))
                            }
                        }
                    }
                }
            }
        }
    }

    func rankColor(for index: Int) -> Color {
        switch index {
        case 0: return .yellow
        case 1: return colorScheme == .dark ? .gray : .gray.opacity(0.8)
        case 2: return .orange
        default: return colorScheme == .dark ? Color.white.opacity(0.55) : Color.black.opacity(0.4)
        }
    }
}

// MARK: - Reusable Views (GlassCard, SectionCard, EmptyStateView, LeaderboardRow)

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
                          ? Color.white.opacity(0.08)
                          : Color.white.opacity(0.6))
                    .overlay(
                        RoundedRectangle(cornerRadius: 22, style: .continuous)
                            .stroke(colorScheme == .dark
                                    ? Color.white.opacity(0.12)
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
                        .background(colorScheme == .dark ? accentBlue.opacity(0.18) : accentBlue.opacity(0.12))
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
                .background(colorScheme == .dark ? Color.white.opacity(0.05) : Color.black.opacity(0.05))
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

private struct LeaderboardRow: View {
    let place: Int
    let name: String
    let rating: Int
    let accent: Color
    var showChevron: Bool = false
    
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
                .background(colorScheme == .dark ? Color.white.opacity(0.05) : Color.black.opacity(0.05))
                .clipShape(Circle())

            Text(name)
                .font(.system(size: 15, weight: .medium))
                .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                .lineLimit(1)

            Spacer(minLength: 0)

            Text("\(rating)")
                .font(.system(size: 15, weight: .bold, design: .rounded))
                .foregroundStyle(accentBlue)
            
            if showChevron {
                Image(systemName: "chevron.right")
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.3) : .black.opacity(0.3))
                    .padding(.leading, 2)
            }
        }
        .padding(.vertical, 2)
    }
}
