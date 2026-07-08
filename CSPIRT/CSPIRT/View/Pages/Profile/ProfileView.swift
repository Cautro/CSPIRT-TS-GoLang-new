import SwiftUI
import UIKit
import PhotosUI

struct ProfileView: View {
    @StateObject private var viewModel = ProfileViewModel()
    @EnvironmentObject private var sessionManager: SessionManager
    @Environment(\.colorScheme) private var colorScheme

    @State private var showLogoutConfirmation = false
    @State private var selectedItem: PhotosPickerItem? = nil

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

                if viewModel.isLoading && viewModel.Name.isEmpty {
                    loadingView
                } else if !viewModel.Name.isEmpty {
                    content
                }
            }
            .task {
                await viewModel.fetchProfile()
            }
            .alert("Выход из профиля", isPresented: $showLogoutConfirmation) {
                Button("Отмена", role: .cancel) {}
                Button("Выйти", role: .destructive) {
                    sessionManager.logout()
                }
            }
        }
    }
}

// MARK: - Content
private extension ProfileView {

    var content: some View {
        ScrollView(.vertical, showsIndicators: false) {
            VStack(spacing: 18) {

                HStack {
                    Spacer()
                    settingButton
                }
                .padding(.top, 16)

                headerSection
                    .onChange(of: selectedItem) { _, newValue in
                        Task {
                            if let data = try? await newValue?.loadTransferable(type: Data.self) {
                                await viewModel.uploadAvatar(imageData: data)
                            }
                        }
                    }

                mainCard
                detailsCard
                paramsCard
            }
            .padding(.horizontal, 36)
            .padding(.bottom, 120)
        }
        .refreshable {
            await viewModel.fetchProfile()
        }
    }

    var loadingView: some View {
        VStack(spacing: 16) {
            ProgressView()
                .scaleEffect(1.5)
                .tint(colorScheme == .dark ? .white : accentBlue)

            Text("Загрузка профиля...")
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.6) : .black.opacity(0.6))
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

// MARK: - Background
private extension ProfileView {

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
}

// MARK: - Header
private extension ProfileView {

    var headerSection: some View {
        VStack(spacing: 16) {
            PhotosPicker(selection: $selectedItem, matching: .images) {
                avatarView
            }
            .buttonStyle(.plain)

            Text(fullNameText)
                .font(.system(size: 24, weight: .bold, design: .rounded))
                .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                .multilineTextAlignment(.center)
        }
    }

    var avatarView: some View {
        ZStack {
            Circle()
                .fill(colorScheme == .dark
                      ? Color.white.opacity(0.07)
                      : Color.black.opacity(0.03))

            if let image = viewModel.avatarImage {
                Image(uiImage: image)
                    .resizable()
                    .scaledToFill()
                    .frame(width: 100, height: 100)
                    .clipShape(Circle())
            } else {
                Image(systemName: "person.crop.circle.fill")
                    .font(.system(size: 76, weight: .thin))
                    .foregroundStyle(colorScheme == .dark
                                     ? .white.opacity(0.7)
                                     : .black.opacity(0.4))
            }
        }
        .frame(width: 116, height: 116)
        .overlay(
            Circle()
                .stroke(colorScheme == .dark
                        ? Color.white.opacity(0.10)
                        : Color.black.opacity(0.10), lineWidth: 1)
        )
        .shadow(color: .black.opacity(colorScheme == .dark ? 0.18 : 0.08),
                radius: 18, x: 0, y: 10)
    }
}

// MARK: - Cards
private extension ProfileView {

    var mainCard: some View {
        GlassCard {
            VStack(alignment: .leading, spacing: 12) {
                Text("Социальный рейтинг")
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.7) : .black.opacity(0.6))

                Text("\(viewModel.Rating)")
                    .font(.system(size: 46, weight: .bold))
                    .foregroundStyle(accentBlue)

                Text("Показатель поведения и активности")
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.5) : .black.opacity(0.5))
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    var detailsCard: some View {
        GlassCard {
            VStack(spacing: 0) {
                ProfileRow(icon: "graduationcap.fill",
                           title: "Класс",
                           value: "\(viewModel.UserClass?.grade ?? 0)\(viewModel.UserClass?.letter ?? "")",
                           accent: accentBlue)

                Divider().opacity(0.15)

                ProfileRow(icon: "at",
                           title: "Логин",
                           value: viewModel.Login.isEmpty ? "–" : "@\(viewModel.Login)",
                           accent: accentBlue)

                Divider().opacity(0.15)

                ProfileRow(icon: "person.fill",
                           title: "Роль",
                           value: UserRole(rawValue: viewModel.Role)?.displayName ?? "User",
                           accent: accentBlue)
            }
        }
    }

    var paramsCard: some View {
        GlassCard {
            VStack(spacing: 0) {
                navRow("square.and.pencil", "К заметкам", NotesPageView())
                Divider().opacity(0.15)
                navRow("doc.text", "К жалобам", ComplaintsPageView())
                Divider().opacity(0.15)
                navRow("person.3", "Мероприятия", EventView())
            }
        }
    }

    func navRow<Destination: View>(_ icon: String, _ title: String, _ destination: Destination) -> some View {
        NavigationLink(destination: destination) {
            HStack {
                ProfileRowWithoutValue(icon: icon, title: title, accent: accentBlue)
                Spacer()
                Image(systemName: "chevron.right")
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.3) : .black.opacity(0.3))
            }
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Settings
private extension ProfileView {

    var settingButton: some View {
        NavigationLink(destination: SettingsView()) {
            Image(systemName: "gearshape.fill")
                .font(.system(size: 22, weight: .semibold))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.8) : .black.opacity(0.6))
                .padding(8)
                .background(
                    Circle().fill(colorScheme == .dark
                                  ? Color.white.opacity(0.07)
                                  : Color.black.opacity(0.05))
                )
        }
    }
}

// MARK: - Helpers
private extension ProfileView {

    var fullNameText: String {
        if let p = viewModel.FullName.first {
            return "\(p.LastName) \(p.Name) \(p.MiddleName)"
        } else {
            return "\(viewModel.Name) \(viewModel.LastName)"
        }
    }
}

// MARK: - UI Kit

private struct GlassCard<Content: View>: View {
    @Environment(\.colorScheme) private var colorScheme
    let content: Content

    init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }

    var body: some View {
        content
            .padding(18)
            .background(
                RoundedRectangle(cornerRadius: 22)
                    .fill(colorScheme == .dark
                          ? Color.white.opacity(0.07)
                          : Color.white.opacity(0.6))
                    .overlay(
                        RoundedRectangle(cornerRadius: 22)
                            .stroke(colorScheme == .dark
                                    ? Color.white.opacity(0.10)
                                    : Color.black.opacity(0.06))
                    )
            )
            .shadow(color: .black.opacity(colorScheme == .dark ? 0.18 : 0.06),
                    radius: 18, x: 0, y: 10)
    }
}

private struct ProfileRow: View {
    let icon: String
    let title: String
    let value: String
    let accent: Color

    var body: some View {
        HStack {
            Image(systemName: icon)
                .foregroundStyle(accent)
                .frame(width: 40, height: 40)
                .background(Color.white.opacity(0.06))
                .cornerRadius(10)

            Text(title)

            Spacer()

            Text(value)
        }
        .frame(height: 56)
    }
}

private struct ProfileRowWithoutValue: View {
    let icon: String
    let title: String
    let accent: Color

    var body: some View {
        HStack {
            Image(systemName: icon)
                .foregroundStyle(accent)
                .frame(width: 40, height: 40)
                .background(Color.white.opacity(0.06))
                .cornerRadius(10)

            Text(title)
        }
        .frame(height: 56)
    }
}
