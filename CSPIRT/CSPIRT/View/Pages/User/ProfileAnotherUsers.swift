import SwiftUI
import UIKit
import PhotosUI

struct UserDetailView: View {
    @EnvironmentObject private var sessionManager: SessionManager
    @StateObject private var viewModel = ProfileViewModel()
    @State private var selectedItem: PhotosPickerItem? = nil
    
    @StateObject private var complaintViewModel = ComplaintViewModel()
    @StateObject private var noteViewModel = NoteViewModel()
    
    @State private var showComplaintSheet = false
    @State private var isShowingAddSheet = false
    
    let userModel: UserModel
    
    @State private var displayedRating: Int = 0
    @State private var showRatingAlert = false
    @State private var ratingInput = ""
    @State private var reasonInput = ""
    
    @Environment(\.colorScheme) private var colorScheme
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }
    
    private var canManageDetails: Bool {
        guard let role = sessionManager.currentUser?.role.lowercased() else { return false }
        return role == "admin" || role == "owner"
    }

    private var isNotUser: Bool {
        guard let role = sessionManager.currentUser?.role.lowercased() else { return false }
        return role == "admin" || role == "owner" || role == "helper"
    }

    var body: some View {
        NavigationStack {
            ZStack {
                backgroundLayer
                    .ignoresSafeArea()
                
                GeometryReader { proxy in
                    ScrollView(.vertical, showsIndicators: false) {
                        VStack(spacing: 20) {
                            headerSection
                                .padding(.top, proxy.safeAreaInsets.top - 100)
                                .padding(.bottom, 8)
                            
                            mainCard
                            
                            detailsCard
                            
                            complaintsBlock
                            
                            if canManageDetails {
                                paramsCard
                            }
                        }
                        .padding(.horizontal, 35)
                        .frame(width: proxy.size.width)
                        .sheet(isPresented: $showComplaintSheet) {
                            AddComplaintSheetView(
                                viewModel: complaintViewModel,
                                targetId: userModel.id ?? 0,
                                targetName: "\(userModel.lastName) \(userModel.name)"
                            ) { text in
                                await complaintViewModel.sendComplaint(
                                    targetId: userModel.id ?? 0,
                                    targetName: "\(userModel.lastName) \(userModel.name)",
                                    createdAt: ISO8601DateFormatter().string(from: Date()),
                                    content: text
                                )
                            }
                        }
                        .sheet(isPresented: $isShowingAddSheet) {
                            AddNoteSheetView(
                                viewModel: noteViewModel,
                                targetId: userModel.id ?? 0,
                                targetName: "\(userModel.lastName) \(userModel.name)"
                            ) 
                        }
                    }
                }
                if showRatingAlert {
                    customRatingAlertOverlay
                }
            }
            .animation(.easeInOut(duration: 0.25), value: showRatingAlert)
            .onAppear {
                displayedRating = userModel.rating
                
                if let user = sessionManager.currentUser {
                    print("🍏 Текущий юзер в деталях: \(user.login), Роль: \(user.role)")
                } else {
                    print("🔴 Внимание! sessionManager.currentUser сейчас NIL. Кнопки не будет!")
                }
            }
        }
    }
}

// MARK: - UI Components

private extension UserDetailView {
    var customRatingAlertOverlay: some View {
        ZStack {
            Color.black.opacity(0.4)
                .ignoresSafeArea()
                .onTapGesture {
                    withAnimation { showRatingAlert = false }
                }
            
            GlassCard {
                VStack(spacing: 16) {
                    Text("Изменение рейтинга")
                        .font(.system(size: 18, weight: .bold, design: .rounded))
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                    
                    Text("Укажите, на сколько баллов изменить рейтинг пользователя @\(userModel.login), и напишите причину.")
                        .font(.system(size: 13, weight: .regular))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.7) : .black.opacity(0.6))
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 4)
                    
                    // Поле ввода очков
                    TextField("", text: $ratingInput, prompt: Text("Количество (например: 15 или -10)").foregroundStyle(colorScheme == .dark ? .white.opacity(0.4) : .black.opacity(0.4)))
                        .keyboardType(.numbersAndPunctuation)
                        .padding(.horizontal, 14)
                        .padding(.vertical, 12)
                        .background(colorScheme == .dark ? Color.white.opacity(0.06) : Color.black.opacity(0.04))
                        .cornerRadius(12)
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                        .overlay(RoundedRectangle(cornerRadius: 12).stroke(colorScheme == .dark ? Color.white.opacity(0.15) : Color.black.opacity(0.1), lineWidth: 1))
                    
                    TextField("", text: $reasonInput, prompt: Text("Причина изменения").foregroundStyle(colorScheme == .dark ? .white.opacity(0.4) : .black.opacity(0.4)))
                        .padding(.horizontal, 14)
                        .padding(.vertical, 12)
                        .background(colorScheme == .dark ? Color.white.opacity(0.06) : Color.black.opacity(0.04))
                        .cornerRadius(12)
                        .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                        .overlay(RoundedRectangle(cornerRadius: 12).stroke(colorScheme == .dark ? Color.white.opacity(0.15) : Color.black.opacity(0.1), lineWidth: 1))
                    
                    HStack(spacing: 12) {
                        Button("Отмена") {
                            withAnimation { showRatingAlert = false }
                        }
                        .font(.system(size: 15, weight: .medium))
                        .foregroundStyle(colorScheme == .dark ? .white.opacity(0.7) : .black.opacity(0.6))
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 12)
                        .background(colorScheme == .dark ? Color.white.opacity(0.1) : Color.black.opacity(0.05))
                        .cornerRadius(12)
                        
                        Button("Сохранить") {
                            if let points = Int(ratingInput) {
                                let pureVM = viewModel
                                Task {
                                    if let newRating = await pureVM.updateRating(
                                        targetLogin: userModel.login,
                                        ratingChange: points,
                                        reason: reasonInput
                                    ) {
                                        self.displayedRating = newRating
                                    }
                                }
                            }
                            withAnimation { showRatingAlert = false }
                        }
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 12)
                        .background(accentBlue)
                        .cornerRadius(12)
                    }
                    .padding(.top, 4)
                }
            }
            .frame(width: 322)
            .shadow(color: .black.opacity(0.35), radius: 20, x: 0, y: 10)
            .transition(.scale(scale: 0.9).combined(with: .opacity))
        }
    }
    
    var complaintsBlock: some View {
        GlassCard {
            VStack(spacing: 0) {
                Button {
                    showComplaintSheet = true
                } label: {
                    ProfileRowWithoutValue(
                        icon: "exclamationmark.bubble.fill",
                        title: "Пожаловаться на ученика",
                        value: ""
                    )
                }
                .buttonStyle(.plain)
                
                if isNotUser {
                    Divider().overlay(colorScheme == .dark ? Color.white.opacity(0.20) : Color.black.opacity(0.1))
                    Button {
                        isShowingAddSheet = true
                    } label: {
                        ProfileRowWithoutValue(
                            icon: "envelope.open.fill",
                            title: "Оставить заметку на ученика",
                            value: ""
                        )
                    }
                }
            }
        }
    }
    
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

    var headerSection: some View {
        VStack(spacing: 14) {
            avatarView
            
            VStack(spacing: 6) {
                Text(fullNameText)
                    .font(.system(size: 26, weight: .bold, design: .rounded))
                    .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
                    .multilineTextAlignment(.center)
            }
        }
    }

    var avatarView: some View {
        ZStack {
            Circle()
                .fill(colorScheme == .dark ? Color.white.opacity(0.10) : Color.black.opacity(0.05))
                .frame(width: 116, height: 116)
                .overlay(
                    Circle()
                        .stroke(colorScheme == .dark ? Color.white.opacity(0.18) : Color.black.opacity(0.08), lineWidth: 1)
                )

            if let base64String = userModel.avatar?.value,
               let data = Data(base64Encoded: base64String),
               let uiImage = UIImage(data: data) {
                Image(uiImage: uiImage)
                    .resizable()
                    .scaledToFill()
                    .frame(width: 100, height: 100)
                    .clipShape(Circle())
            } else {
                Image(systemName: "person.crop.circle.fill")
                    .font(.system(size: 76, weight: .regular))
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.85) : .black.opacity(0.6))
            }
        }
        .shadow(color: .black.opacity(colorScheme == .dark ? 0.25 : 0.1), radius: 12, x: 0, y: 6)
    }

    var mainCard: some View {
        GlassCard {
            VStack(alignment: .leading, spacing: 14) {
                Text("Социальный рейтинг")
                    .font(.system(size: 15, weight: .medium))
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.65) : .black.opacity(0.6))

                HStack(alignment: .center, spacing: 10) {
                    Text("\(displayedRating)")
                        .font(.system(size: 42, weight: .bold, design: .rounded))
                        .foregroundStyle(accentBlue)
                    
                    Spacer()
                    
                    if canManageDetails {
                        Button {
                            ratingInput = ""
                            reasonInput = ""
                            withAnimation { showRatingAlert = true }
                        } label: {
                            HStack(spacing: 6) {
                                Image(systemName: "plusminus.circle.fill")
                                Text("Изменить")
                            }
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundStyle(accentBlue)
                            .padding(.horizontal, 14)
                            .padding(.vertical, 8)
                            .background(colorScheme == .dark ? accentBlue.opacity(0.25) : accentBlue.opacity(0.15))
                            .cornerRadius(12)
                            .overlay(
                                RoundedRectangle(cornerRadius: 12)
                                    .stroke(colorScheme == .dark ? accentBlue.opacity(0.4) : accentBlue.opacity(0.3), lineWidth: 1)
                            )
                        }
                    }
                }

                Text("Показатель поведения, активности и вклада в школьную жизнь.")
                    .font(.system(size: 13, weight: .regular))
                    .foregroundStyle(colorScheme == .dark ? .white.opacity(0.50) : .black.opacity(0.5))
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    var detailsCard: some View {
        GlassCard {
            VStack(spacing: 0) {
                ProfileRow(icon: "graduationcap.fill", title: "Класс", value: userModel.className)
                Divider().overlay(colorScheme == .dark ? Color.white.opacity(0.20) : Color.black.opacity(0.1))
                ProfileRow(icon: "at", title: "Логин", value: "@\(userModel.login)")
                Divider().overlay(colorScheme == .dark ? Color.white.opacity(0.20) : Color.black.opacity(0.1))
                ProfileRow(icon: "person.fill", title: "Роль", value: userModel.role.displayName)
            }
        }
    }
    
    var paramsCard: some View {
        GlassCard {
            VStack {
                HStack {
                    NavigationLink(destination: NotesPageView(
                        targetUserId: userModel.id,
                        targetUserName: "\(userModel.lastName) \(userModel.name)".trimmingCharacters(in: .whitespaces)
                    )) {
                        ProfileRowWithoutValue(icon: "square.and.pencil", title: "К заметкам ученика", value: "")
                        Image(systemName: "chevron.right")
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundStyle(colorScheme == .dark ? .white.opacity(0.4) : .black.opacity(0.3))
                    }
                }
                Divider().overlay(colorScheme == .dark ? Color.white.opacity(0.20) : Color.black.opacity(0.1))
                HStack {
                    NavigationLink(destination: ComplaintsPageView(
                        targetUserId: userModel.id,
                        targetUserName: "\(userModel.lastName) \(userModel.name)".trimmingCharacters(in: .whitespaces)
                    )) {
                        ProfileRowWithoutValue(icon: "doc.text", title: "К жалобам ученика", value: "")
                        Image(systemName: "chevron.right")
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundStyle(colorScheme == .dark ? .white.opacity(0.4) : .black.opacity(0.3))
                    }
                }
                Divider().overlay(colorScheme == .dark ? Color.white.opacity(0.20) : Color.black.opacity(0.1))
                HStack {
                    NavigationLink(destination: EventView()) {
                        ProfileRowWithoutValue(icon: "person.3", title: "К мероприятиям", value: "")
                        Image(systemName: "chevron.right")
                            .font(.system(size: 14, weight: .semibold))
                            .foregroundStyle(colorScheme == .dark ? .white.opacity(0.4) : .black.opacity(0.3))
                    }
                }
            }
        }
    }
    
    var fullNameText: String {
        let composed = "\(userModel.lastName) \(userModel.name) \(userModel.fullName.first?.MiddleName ?? "")"
            .trimmingCharacters(in: .whitespaces)
        return composed.isEmpty ? "Профиль" : composed
    }
}

// MARK: - Вспомогательные View структуры
private struct GlassCard<Content: View>: View {
    @Environment(\.colorScheme) private var colorScheme
    @ViewBuilder let content: Content

    var body: some View {
        content
            .padding(18)
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

private struct ProfileRow: View {
    let icon: String
    let title: String
    let value: String
    
    @Environment(\.colorScheme) private var colorScheme
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }

    var body: some View {
        HStack(spacing: 14) {
            ZStack {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .fill(colorScheme == .dark ? accentBlue.opacity(0.18) : accentBlue.opacity(0.12))

                Image(systemName: icon)
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(accentBlue)
            }
            .frame(width: 42, height: 42)

            Text(title)
                .font(.system(size: 15, weight: .medium))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.70) : .black.opacity(0.6))

            Spacer()

            Text(value.isEmpty ? "–" : value)
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
        }
        .frame(height: 56)
    }
}

private struct ProfileRowWithoutValue: View {
    let icon: String
    let title: String
    let value: String
    
    @Environment(\.colorScheme) private var colorScheme
    
    private var accentBlue: Color {
        colorScheme == .dark
        ? Color(red: 0.0, green: 0.65, blue: 1.0)
        : Color(red: 0.22, green: 0.42, blue: 0.88)
    }

    var body: some View {
        HStack(spacing: 14) {
            ZStack {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .fill(colorScheme == .dark ? accentBlue.opacity(0.18) : accentBlue.opacity(0.12))

                Image(systemName: icon)
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(accentBlue)
            }
            .frame(width: 42, height: 42)

            Text(title)
                .font(.system(size: 15, weight: .medium))
                .foregroundStyle(colorScheme == .dark ? .white.opacity(0.70) : .black.opacity(0.6))

            Spacer()

            Text(value.isEmpty ? "" : value)
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(colorScheme == .dark ? .white : .black.opacity(0.9))
        }
        .frame(height: 56)
    }
}
