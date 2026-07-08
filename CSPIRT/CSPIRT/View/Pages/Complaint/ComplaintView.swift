import SwiftUI

struct ComplaintsPageView: View {
    let targetUserId: Int?
    let targetUserName: String?
    
    @StateObject private var viewModel = ComplaintViewModel()
    @EnvironmentObject private var sessionManager: SessionManager
    
    @State private var isShowingAddSheet = false
    
    init(targetUserId: Int? = nil, targetUserName: String? = nil) {
        self.targetUserId = targetUserId
        self.targetUserName = targetUserName
    }
    
    var currentUserRole: String {
        viewModel.user?.role.rawValue ?? "User"
    }
    
    private var isNotUser: Bool {
        guard let role = sessionManager.currentUser?.role.lowercased() else { return false }
        return role == "helper" || role == "admin" || role == "owner"
    }

    var body: some View {
        ZStack {
            backgroundLayer
                .ignoresSafeArea()
            
            if viewModel.isLoading && viewModel.complaints.isEmpty {
                VStack(spacing: 16) {
                    ProgressView()
                        .scaleEffect(1.5)
                        .tint(.white)
                    Text("Загрузка жалоб...")
                        .font(.system(size: 16, weight: .medium, design: .rounded))
                        .foregroundStyle(.white.opacity(0.6))
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                
            } else if let errMsg = viewModel.errMsg, viewModel.complaints.isEmpty {
                VStack(spacing: 20) {
                    Image(systemName: "exclamationmark.triangle")
                        .font(.system(size: 40, weight: .medium))
                        .foregroundStyle(.white.opacity(0.7))
                    
                    Text(errMsg)
                        .font(.system(size: 16, weight: .medium, design: .rounded))
                        .foregroundStyle(.white.opacity(0.9))
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 32)
                    
                    Button {
                        Task { await viewModel.fetchComplaints(for: targetUserId) }
                    } label: {
                        Text("Повторить попытку")
                            .font(.system(size: 15, weight: .semibold, design: .rounded))
                            .foregroundStyle(.white)
                            .padding(.horizontal, 24)
                            .padding(.vertical, 12)
                            .background(Color(red: 0.0, green: 0.55, blue: 1.0))
                            .cornerRadius(12)
                    }
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                
            } else {
                ScrollView(.vertical, showsIndicators: false) {
                    VStack(spacing: 16) {
                        headerView
                            .padding(.horizontal, 15)
                            .padding(.top, 10)

                        complaintsListSection
                            .padding(.horizontal, 15)
                    }
                    .padding(.horizontal, 16)
                    .padding(.top, 16)
                    .padding(.bottom, 24)
                }
                .refreshable {
                    await viewModel.fetchComplaints(for: targetUserId)
                }
            }
        }
        .task {
            await viewModel.fetchComplaints(for: targetUserId)
        }
    }
}
// MARK: - UI Components

private extension ComplaintsPageView {
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
                            Color.black.opacity(0.35),
                            Color.black.opacity(0.70),
                            Color.black.opacity(0.92)
                        ],
                        startPoint: .top,
                        endPoint: .bottom
                    )
                }

            Color.black.opacity(0.12)
        }
    }

    var headerView: some View {
        GlassCard {
            HStack(alignment: .top, spacing: 12) {
                VStack(alignment: .leading, spacing: 8) {
                    Text(targetUserName != nil ? "Жалобы" : "Мои жалобы")
                        .font(.system(size: 28, weight: .bold, design: .rounded))
                        .foregroundStyle(.white)

                    Text(targetUserName != nil ? "Жалобы на ученика: \(targetUserName!)" : "История зафиксированных нарушений и обращений")
                        .font(.system(size: 14, weight: .regular))
                        .foregroundStyle(.white.opacity(0.62))
                        .fixedSize(horizontal: false, vertical: true)
                }

                Spacer(minLength: 0)

                Image(systemName: "exclamationmark.bubble.fill")
                    .font(.system(size: 18, weight: .semibold))
                    .foregroundStyle(.red)
                    .frame(width: 40, height: 40)
                    .background(Color.white.opacity(0.06))
                    .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            }
        }
    }

    var complaintsListSection: some View {
        VStack(spacing: 12) {
            if viewModel.complaints.isEmpty {
                GlassCard {
                    EmptyStateView(
                        icon: "checkmark.shield",
                        title: "Жалоб пока нет",
                        subtitle: targetUserName != nil ? "На этого ученика не поступало жалоб." : "Прекрасно! На вас не поступало жалоб от других участников."
                    )
                }
            } else {
                ForEach(viewModel.complaints, id: \.id) { complaint in
                    ComplaintRowCard(complaint: complaint, currentUserRole: currentUserRole,
                        onDelete: {
                            Task {
                                _ = await viewModel.deleteComplaint(complaintId: complaint.id)
                            }
                        }
                    )
                }
            }
        }
    }
}

// MARK: - Private Row Card

private struct ComplaintRowCard: View {
    let complaint: ComplaintModel
    let currentUserRole: String
    let onDelete: () -> Void
    
    private var canSeeAuthor: Bool {
        ["Admin", "Owner", "SystemAdmin"].contains(currentUserRole)
    }
    
    var body: some View {
        GlassCard(padding: 0) {
            HStack(alignment: .top, spacing: 0) {
                RoundedRectangle(cornerRadius: 0)
                    .fill(Color.red)
                    .frame(width: 5)
                
                VStack(alignment: .leading, spacing: 12) {
                    HStack(alignment: .center) {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(canSeeAuthor ? complaint.authorName : "Отправитель скрыт")
                                .font(.system(size: 15, weight: .semibold, design: .rounded))
                                .foregroundStyle(canSeeAuthor ? .white : .white.opacity(0.6))
                            
                            if !complaint.targetName.isEmpty {
                                Text("На кого: \(complaint.targetName)")
                                    .font(.system(size: 12, weight: .medium))
                                    .foregroundStyle(.white.opacity(0.45))
                            }
                        }
                        
                        Spacer()
                        
                        Text(formatDate(complaint.createdAt))
                            .font(.system(size: 12, weight: .regular))
                            .foregroundStyle(.white.opacity(0.5))
                    }
                    
                    Text(complaint.content)
                        .font(.system(size: 14))
                        .foregroundStyle(.white.opacity(0.88))
                        .lineSpacing(4)
                        .fixedSize(horizontal: false, vertical: true)
                }
                .padding(16)
            }
        }
        .clipShape(RoundedRectangle(cornerRadius: 22, style: .continuous))
        .contextMenu {
            Button {
                UIPasteboard.general.string = complaint.content
            } label: {
                Label("Скопировать текст", systemImage: "doc.on.doc")
            }
            
            Button(role: .destructive) {
                onDelete()
            } label: {
                Label("Удалить", systemImage: "trash")
            }
        }
    }
    
    private func formatDate(_ isoString: String) -> String {
        return isoString.prefix(10).replacingOccurrences(of: "-", with: ".")
    }
}

// MARK: - Reusable UI Components

private struct GlassCard<Content: View>: View {
    var padding: CGFloat = 16
    @ViewBuilder let content: Content

    var body: some View {
        content
            .padding(padding)
            .background(
                RoundedRectangle(cornerRadius: 22, style: .continuous)
                    .fill(Color.white.opacity(0.07))
                    .overlay(
                        RoundedRectangle(cornerRadius: 22, style: .continuous)
                            .stroke(Color.white.opacity(0.10), lineWidth: 1)
                    )
                    .shadow(color: .black.opacity(0.18), radius: 18, x: 0, y: 10)
            )
    }
}

private struct EmptyStateView: View {
    let icon: String
    let title: String
    let subtitle: String

    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: icon)
                .font(.system(size: 20, weight: .semibold))
                .foregroundStyle(.white.opacity(0.7))
                .frame(width: 44, height: 44)
                .background(Color.white.opacity(0.05))
                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))

            Text(title)
                .font(.system(size: 15, weight: .semibold))
                .foregroundStyle(.white)

            Text(subtitle)
                .font(.system(size: 13))
                .foregroundStyle(.white.opacity(0.5))
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 12)
    }
}
