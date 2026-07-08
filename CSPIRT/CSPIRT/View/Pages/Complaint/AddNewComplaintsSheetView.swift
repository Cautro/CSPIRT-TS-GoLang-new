import SwiftUI
import Foundation
import Combine

struct AddComplaintSheetView: View {
    @ObservedObject var viewModel: ComplaintViewModel

    let targetId: Int
    let targetName: String

    let onSend: (String) async -> Bool

    @Environment(\.dismiss) private var dismiss

    @State private var complaintText = ""
    @State private var showAlert = false

    var isFormValid: Bool {
        !complaintText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }
    
    var body: some View {
        NavigationStack {
            ZStack {
                Color.black.opacity(0.95).ignoresSafeArea()
                
                VStack(spacing: 20) {
                    Text("Новое обращение")
                        .font(.system(size: 24, weight: .bold, design: .rounded))
                        .foregroundStyle(.white)
                        .padding(.top, 10)
                    
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Жалоба на")
                            .foregroundStyle(.white.opacity(0.6))

                        Text(targetName)
                            .font(.title3.bold())
                            .foregroundStyle(.white)
                    }
                    
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Суть жалобы")
                            .font(.system(size: 14, weight: .medium))
                            .foregroundStyle(.white.opacity(0.6))
                        
                        TextEditor(text: $complaintText)
                            .frame(height: 150)
                            .padding(8)
                            .background(Color.white.opacity(0.06))
                            .cornerRadius(12)
                            .foregroundStyle(.white)
                            .scrollContentBackground(.hidden)
                            .overlay(
                                RoundedRectangle(cornerRadius: 12)
                                    .stroke(Color.white.opacity(0.1), lineWidth: 1)
                            )
                    }
                    
                    Spacer()
                    
                    Button {
                        Task {
                            let success = await viewModel.sendComplaint(
                                targetId: targetId,
                                targetName: targetName,
                                createdAt: ISO8601DateFormatter().string(from: Date()),
                                content: complaintText
                            )

                            if success {
                                dismiss()
                            } else {
                                showAlert = true
                            }
                        }
                    } label: {
                        HStack {
                            if viewModel.isLoading {
                                ProgressView().tint(.white)
                            } else {
                                Text("Отправить жалобу")
                                    .font(.system(size: 16, weight: .bold, design: .rounded))
                            }
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 16)
                        .background(isFormValid ? Color.red.opacity(0.8) : Color.white.opacity(0.1))
                        .foregroundStyle(isFormValid ? .white : .white.opacity(0.3))
                        .cornerRadius(14)
                    }
                    .disabled(!isFormValid || viewModel.isLoading)
                }
                .padding(24)
            }
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button("Отмена") { dismiss() }
                        .foregroundStyle(.white.opacity(0.6))
                }
            }
            .alert("Ошибка", isPresented: $showAlert) {
                Button("OK", role: .cancel) {}
            } message: {
                Text(viewModel.errMsg ?? "Что-то пошло не так")
            }
        }
    }
}
