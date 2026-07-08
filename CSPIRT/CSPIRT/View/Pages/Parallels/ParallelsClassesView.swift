import SwiftUI

struct ParallelsClassesView: View {
    let parallel: ParallelModel
    let classes: [ClassesModel]
    let bestClass: ClassesModel?
    
    @Environment(\.colorScheme) private var colorScheme
    
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
                
                ScrollView(.vertical, showsIndicators: false) {
                    VStack(spacing: 16) {
                        headerView
                        bestClassSection
                        classesSection
                    }
                    .padding(.horizontal, 16)
                    .padding(.top, 16)
                    .padding(.bottom, 24)
                }
            }
        }
    }
}

// MARK: - UI Components

private extension ParallelsClassesView {
    var backgroundLayer: some View {
        ZStack {
            Image("school_background")
                .resizable()
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

    var headerView: some View {
        GlassCard {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 8) {
                    Text(parallel.name)
                        .font(.system(size: 28, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

                    Text("Все классы этой параллели и их общий рейтинг")
                        .font(.system(size: 14, weight: .regular))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.62) : .black.opacity(0.6))
                        .fixedSize(horizontal: false, vertical: true)
                }

                Spacer(minLength: 0)

                Image(systemName: "person.3.fill")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundStyle(accentBlue)
                    .frame(width: 40, height: 40)
                    .background(colorScheme == .dark ? Color.white.opacity(0.06) : Color.black.opacity(0.05))
                    .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            }
        }
    }

    var bestClassSection: some View {
        SectionCard(title: "Лучший класс", icon: "crown.fill") {
            if let bestClass {
                NavigationLink(destination: ClassDetailView(classModel: bestClass)) {
                    BestClassCard(bestClass: bestClass)
                }
                .buttonStyle(PlainButtonStyle())
            } else {
                EmptyStateView(
                    icon: "crown",
                    title: "Лучший класс не найден",
                    subtitle: "Пока нет данных для расчёта."
                )
            }
        }
    }

    var classesSection: some View {
        SectionCard(title: "Классы в параллели", icon: "person.2.fill") {
            if classes.isEmpty {
                EmptyStateView(
                    icon: "person.2.slash.fill",
                    title: "Классы не найдены",
                    subtitle: "В этой параллели пока нет классов."
                )
            } else {
                VStack(spacing: 10) {
                    ForEach(classes, id: \.id) { classModel in
                        NavigationLink(destination: ClassDetailView(classModel: classModel)) {
                            ClassCardRow(classModel: classModel)
                        }
                        .buttonStyle(PlainButtonStyle())
                    }
                }
            }
        }
    }
}

// MARK: - Private Views

private struct BestClassCard: View {
    let bestClass: ClassesModel
    @Environment(\.colorScheme) private var colorScheme
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 4) {
                    Text("\(bestClass.grade)\(bestClass.letter)")
                        .font(.system(size: 20, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

                    Text("\(bestClass.grade)\(bestClass.letter) • Кл. руководитель: \(bestClass.teacher.lastName) \(bestClass.teacher.name.prefix(1)).")
                        .font(.system(size: 13))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.55) : .black.opacity(0.6))
                }

                Spacer(minLength: 0)

                HStack(spacing: 4) {
                    Text("\(bestClass.classTotalRating)")
                        .font(.system(size: 16, weight: .bold, design: .rounded))
                    Image(systemName: "star.fill")
                        .font(.system(size: 11))
                }
                .foregroundStyle(accentBlue)
                .padding(.horizontal, 10)
                .padding(.vertical, 6)
                .background(colorScheme == .dark ? Color.white.opacity(0.1) : accentBlue.opacity(0.1))
                .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            }

            Divider()
                .overlay(colorScheme == .dark ? Color.white.opacity(0.08) : Color.black.opacity(0.08))

            HStack(spacing: 12) {
                infoChip(label: "Учеников", value: "\(bestClass.members.count)")
                infoChip(label: "Рейтинг класса", value: "\(bestClass.classTotalRating)")
                infoChip(label: "Рейтинг ученика", value: "\(bestClass.userTotalRating)")
            }
        }
        .padding(14)
        .background(colorScheme == .dark ? Color.white.opacity(0.05) : Color.black.opacity(0.02))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(Color.orange.opacity(colorScheme == .dark ? 0.18 : 0.4), lineWidth: 1)
        )
    }

    private func infoChip(label: String, value: String) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(label)
                .font(.system(size: 11, weight: .medium))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.45) : .black.opacity(0.5))

            Text(value)
                .font(.system(size: 14, weight: .semibold, design: .rounded))
                .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(10)
        .background(colorScheme == .dark ? Color.white.opacity(0.04) : Color.black.opacity(0.04))
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
    }
}

private struct ClassCardRow: View {
    let classModel: ClassesModel
    @Environment(\.colorScheme) private var colorScheme
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text("\(classModel.grade)-\(classModel.letter)")
                        .font(.system(size: 20, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                    
                    Text("Кл. рук: \(classModel.teacher.lastName) \(classModel.teacher.name.prefix(1)).")
                        .font(.system(size: 13, weight: .regular))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.6) : .black.opacity(0.6))
                }
                
                Spacer()
                
                Image(systemName: "chevron.right")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.3) : .black.opacity(0.3))
            }
            
            Divider()
                .overlay(colorScheme == .dark ? Color.white.opacity(0.08) : Color.black.opacity(0.08))
            
            HStack(spacing: 16) {
                HStack(spacing: 6) {
                    Image(systemName: "star.fill")
                        .font(.system(size: 12))
                        .foregroundStyle(colorScheme == .dark ? .yellow : .orange)
                    Text("\(classModel.classTotalRating)")
                        .font(.system(size: 13, weight: .semibold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                }
                
                HStack(spacing: 6) {
                    Image(systemName: "person.2.fill")
                        .font(.system(size: 12))
                        .foregroundStyle(accentBlue)
                    Text("\(classModel.members.count) уч.")
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.7) : .black.opacity(0.7))
                }
            }
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 18, style: .continuous)
                .fill(colorScheme == .dark ? Color.white.opacity(0.06) : Color.black.opacity(0.03))
                .overlay(
                    RoundedRectangle(cornerRadius: 18, style: .continuous)
                        .stroke(colorScheme == .dark ? Color.white.opacity(0.08) : Color.black.opacity(0.05), lineWidth: 1)
                )
        )
    }
}

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
