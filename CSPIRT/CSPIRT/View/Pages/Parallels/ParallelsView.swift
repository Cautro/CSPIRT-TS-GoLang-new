import SwiftUI

struct ParallelsView: View {
    @StateObject private var viewModel = ParallelsViewModel()
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

            if viewModel.isLoading && viewModel.parallels.isEmpty {
                loadingView
            } else if let errMsg = viewModel.errMsg, viewModel.parallels.isEmpty {
                errorView(errMsg: errMsg)
            } else {
                ScrollView(.vertical, showsIndicators: false) {
                    VStack(spacing: 16) {
                        headerView
                            .padding(.horizontal, 15)
                            .padding(.top, 10)

                        //summaryStrip
                        //  .padding(.horizontal, 15)

                        parallelsSection
                            .padding(.horizontal, 15)
                    }
                    .padding(.horizontal, 16)
                    .padding(.top, 16)
                    .padding(.bottom, 24)
                }
                .refreshable {
                    await viewModel.fetchParallels()
                }
            }
        }
        .sheet(isPresented: $viewModel.isShowingParallelDetails) {
            if let parallel = viewModel.selectedParallel {
                ParallelsClassesView(
                    parallel: parallel,
                    classes: viewModel.selectedParallelClasses,
                    bestClass: viewModel.selectedBestClass
                )
                .presentationDetents([.large])
                .presentationDragIndicator(.visible)
            } else {
                Text("Нет выбранной параллели")
                    .font(.system(size: 16, weight: .medium))
                    .foregroundStyle(colorScheme == .dark ? .white : .black)
                    .padding()
            }
        }
        .task {
            await viewModel.fetchParallels()
        }
    }
}

// MARK: - UI Components

private extension ParallelsView {
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

            Text("Загрузка параллелей...")
                .font(.system(size: 16, weight: .medium, design: .rounded))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.6) : .black.opacity(0.6))
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    func errorView(errMsg: String) -> some View {
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
                Task { await viewModel.fetchParallels() }
            } label: {
                Text("Повторить попытку")
                    .font(.system(size: 15, weight: .semibold, design: .rounded))
                    .foregroundStyle(.white) // Оставляем белым на цветной кнопке
                    .padding(.horizontal, 24)
                    .padding(.vertical, 12)
                    .background(accentBlue)
                    .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
            }
            .shadow(color: accentBlue.opacity(0.3), radius: 8, x: 0, y: 4)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    var headerView: some View {
        GlassCard {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Параллели")
                        .font(.system(size: 28, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))

                    Text("Нажми на параллель, чтобы открыть классы и лучший класс")
                        .font(.system(size: 14, weight: .regular))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.62) : .black.opacity(0.6))
                        .fixedSize(horizontal: false, vertical: true)
                }

                Spacer(minLength: 0)

                Image(systemName: "square.stack.3d.up.fill")
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
            statCard(
                title: "Параллелей",
                value: "\(viewModel.parallels.count)",
                icon: "square.grid.2x2.fill"
            )

            statCard(
                title: "Классов",
                value: "\(viewModel.classesInParallel.count)",
                icon: "person.3.fill"
            )

            statCard(
                title: "Лучший",
                value: viewModel.bestClass?.name ?? "—",
                icon: "trophy.fill"
            )
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
                        .font(.system(size: 18, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                        .lineLimit(1)
                        .minimumScaleFactor(0.75)

                    Text(title)
                        .font(.system(size: 12, weight: .medium))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.52) : .black.opacity(0.5))
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    var parallelsSection: some View {
        SectionCard(title: "Доступные параллели", icon: "list.bullet.rectangle.portrait.fill") {
            if viewModel.parallels.isEmpty {
                EmptyStateView(
                    icon: "folder.badge.questionmark",
                    title: "Нет данных",
                    subtitle: "Список параллелей пока пуст."
                )
            } else {
                VStack(spacing: 12) {
                    ForEach(viewModel.parallels.sorted { $0.id < $1.id }, id: \.id) { parallel in
                        Button {
                            Task {
                                await viewModel.selectParallel(parallel)
                            }
                        } label: {
                            ParallelRowCard(parallel: parallel)
                        }
                        .buttonStyle(.plain)
                    }
                }
            }
        }
    }
}

// MARK: - Reusable Views

private struct ParallelRowCard: View {
    let parallel: ParallelModel
    @Environment(\.colorScheme) private var colorScheme
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }

    var body: some View {
        HStack(spacing: 14) {
            Text("\(parallel.id)")
                .font(.system(size: 18, weight: .bold, design: .rounded))
                .foregroundStyle(.white) // Оставляем текст белым, т.к. фон акцентный
                .frame(width: 44, height: 44)
                .background(accentBlue.opacity(0.9))
                .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))

            VStack(alignment: .leading, spacing: 4) {
                Text(parallel.name)
                    .font(.system(size: 16, weight: .semibold, design: .rounded))
                    .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                    .lineLimit(1)

                Text("Классов: \(parallel.classesIds.count)")
                    .font(.system(size: 13))
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.5) : .black.opacity(0.6))
                    .lineLimit(1)
            }

            Spacer(minLength: 0)

            Image(systemName: "chevron.right")
                .font(.system(size: 14, weight: .semibold))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.3) : .black.opacity(0.3))
        }
        .padding(12)
        .background(colorScheme == .dark ? Color.white.opacity(0.05) : Color.black.opacity(0.03))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 16, style: .continuous)
                .stroke(colorScheme == .dark ? Color.white.opacity(0.08) : Color.black.opacity(0.05), lineWidth: 1)
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
