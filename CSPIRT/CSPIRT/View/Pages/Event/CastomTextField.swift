import SwiftUI

struct CustomTextField: View {
    let placeholder: String
    @Binding var text: String
    var isEditor: Bool = false
    var keyboardType: UIKeyboardType = .default
    
    var body: some View {
        ZStack(alignment: .topLeading) {
            if isEditor {
                TextEditor(text: $text)
                    .font(.system(size: 16))
                    .frame(minHeight: 100)
                    .scrollContentBackground(.hidden)
                    .padding(12)
                    .background(Color.white.opacity(0.06))
                    .cornerRadius(12)
                    .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.white.opacity(0.08), lineWidth: 1))
                    .foregroundStyle(.white)
                
                if text.isEmpty {
                    Text(placeholder)
                        .foregroundStyle(.white.opacity(0.4))
                        .padding(.horizontal, 16)
                        .padding(.vertical, 20)
                        .allowsHitTesting(false)
                }
            } else {
                TextField("", text: $text)
                    .keyboardType(keyboardType)
                    .font(.system(size: 16))
                    .padding(16)
                    .background(Color.white.opacity(0.06))
                    .cornerRadius(12)
                    .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.white.opacity(0.08), lineWidth: 1))
                    .foregroundStyle(.white)
                    .overlay(
                        Text(placeholder)
                            .foregroundStyle(.white.opacity(0.4))
                            .padding(.horizontal, 16)
                            .opacity(text.isEmpty ? 1 : 0)
                            .allowsHitTesting(false)
                        , alignment: .leading
                    )
            }
        }
    }
}

struct AddPlayerSheet: View {
    @ObservedObject var viewModel: EventViewModel
    let eventId: Int
    @Binding var currentPlayers: [SafeUserModel] // ИСПРАВЛЕНО: Теперь это @Binding
    let onComplete: () -> Void
    @Environment(\.dismiss) private var dismiss
    
    @State private var searchText = ""
    @State private var selectedUserIds: Set<Int> = []
    @State private var isSubmitting = false
    
    var availableUsers: [SafeUserModel] {
        let currentIds = Set(currentPlayers.map { $0.id })
        let filtered = viewModel.allUsers.filter { !currentIds.contains($0.id) }
        
        if searchText.isEmpty {
            return filtered
        } else {
            let lowercasedSearch = searchText.lowercased()
            return filtered.filter { user in
                let nameMatches = user.name.lowercased().contains(lowercasedSearch)
                let loginMatches = user.login.lowercased().contains(lowercasedSearch)
                let fullNameMatches = user.fullName.contains { fullNameObj in
                    fullNameObj.Name.lowercased().contains(lowercasedSearch) ||
                    fullNameObj.LastName.lowercased().contains(lowercasedSearch)
                }
                
                return nameMatches || loginMatches || fullNameMatches
            }
        }
    }
    
    var body: some View {
        NavigationView {
            ZStack {
                Color(red: 0.05, green: 0.05, blue: 0.07).ignoresSafeArea()
                
                VStack(spacing: 0) {
                    TextField("Поиск по имени или логину", text: $searchText)
                        .padding(12)
                        .background(Color.white.opacity(0.1))
                        .cornerRadius(10)
                        .padding()
                        .foregroundStyle(.white)
                        .colorScheme(.dark)
                    
                    List {
                        ForEach(availableUsers) { user in
                            let isSelected = selectedUserIds.contains(user.id)
                            
                            Button {
                                if isSelected {
                                    selectedUserIds.remove(user.id)
                                } else {
                                    selectedUserIds.insert(user.id)
                                }
                            } label: {
                                HStack {
                                    VStack(alignment: .leading) {
                                        // ИСПРАВЛЕНО: Корректное отображение Фамилии и Имени вместо массива
                                        Text(user.fullName.first.map { "\($0.LastName) \($0.Name)" } ?? user.name)
                                            .foregroundStyle(.white)
                                        Text(user.login)
                                            .font(.caption)
                                            .foregroundStyle(.gray)
                                    }
                                    Spacer()
                                    if isSelected {
                                        Image(systemName: "checkmark.circle.fill")
                                            .foregroundColor(.blue)
                                    } else {
                                        Image(systemName: "circle")
                                            .foregroundColor(.gray)
                                    }
                                }
                            }
                            .listRowBackground(Color.white.opacity(0.06))
                        }
                    }
                    .scrollContentBackground(.hidden)
                    
                    Button {
                        addSelectedPlayers()
                    } label: {
                        Text(isSubmitting ? "Добавление..." : "Добавить (\(selectedUserIds.count))")
                            .font(.system(size: 16, weight: .bold))
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(selectedUserIds.isEmpty ? Color.gray : Color.blue)
                            .foregroundStyle(.white)
                            .cornerRadius(12)
                    }
                    .disabled(selectedUserIds.isEmpty || isSubmitting)
                    .padding()
                }
            }
            .navigationTitle("Новые участники")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Отмена") { dismiss() }
                }
            }
        }
    }
    
    private func addSelectedPlayers() {
        isSubmitting = true
        Task {
            do {
                try await viewModel.addPlayersToEvent(eventId: eventId, playerIds: Array(selectedUserIds))
                onComplete()
                dismiss()
            } catch {
                print("Failed to add players: \(error)")
                isSubmitting = false
            }
        }
    }
}
