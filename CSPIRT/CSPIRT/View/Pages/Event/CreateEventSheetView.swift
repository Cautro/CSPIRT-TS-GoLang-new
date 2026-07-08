import SwiftUI

struct CreateEventSheet: View {
    @ObservedObject var viewModel: EventViewModel
    @Environment(\.dismiss) private var dismiss
    
    @State private var title: String = ""
    @State private var description: String = ""
    @State private var rewardString: String = ""
    @State private var startDate: Date = Date()
    @State private var selectedClasses: Set<Int> = []
    
    @State private var isSubmitting = false
    @State private var errorMessage: String?
    
    var body: some View {
        NavigationView {
            ZStack {
                Color(red: 0.05, green: 0.05, blue: 0.07).ignoresSafeArea()
                
                ScrollView {
                    VStack(spacing: 20) {
                        if let error = errorMessage {
                            Text(error)
                                .foregroundColor(.red)
                                .font(.system(size: 14, weight: .medium))
                                .padding()
                                .background(Color.red.opacity(0.1))
                                .cornerRadius(10)
                        }
                        
                        VStack(alignment: .leading, spacing: 16) {
                            Text("ОСНОВНАЯ ИНФОРМАЦИЯ")
                                .font(.system(size: 12, weight: .bold, design: .rounded))
                                .foregroundStyle(.white.opacity(0.4))
                                .tracking(1.5)
                            
                            CustomTextField(placeholder: "Название мероприятия", text: $title)
                            CustomTextField(placeholder: "Описание", text: $description, isEditor: true)
                            CustomTextField(placeholder: "Награда (очки рейтинга)", text: $rewardString, keyboardType: .numberPad)
                            
                            HStack {
                                Text("Дата начала")
                                    .foregroundStyle(.white.opacity(0.8))
                                Spacer()
                                DatePicker("", selection: $startDate, displayedComponents: [.date, .hourAndMinute])
                                    .labelsHidden()
                                    .colorScheme(.dark)
                            }
                            .padding()
                            .background(Color.white.opacity(0.06))
                            .cornerRadius(12)
                            .overlay(RoundedRectangle(cornerRadius: 12).stroke(Color.white.opacity(0.08), lineWidth: 1))
                        }
                        
                        VStack(alignment: .leading, spacing: 16) {
                            Text("ДОСТУПНО КЛАССАМ")
                                .font(.system(size: 12, weight: .bold, design: .rounded))
                                .foregroundStyle(.white.opacity(0.4))
                                .tracking(1.5)
                            
                            if viewModel.classesInEvents.isEmpty {
                                Text("Нет доступных классов")
                                    .foregroundColor(.white.opacity(0.5))
                            } else {
                                LazyVGrid(columns: [GridItem(.adaptive(minimum: 60))], spacing: 10) {
                                    ForEach(viewModel.classesInEvents, id: \.id) { cls in
                                        let isSelected = selectedClasses.contains(cls.id)
                                        Button {
                                            if isSelected {
                                                selectedClasses.remove(cls.id)
                                            } else {
                                                selectedClasses.insert(cls.id)
                                            }
                                        } label: {
                                            Text("\(cls.grade)\(cls.letter)")
                                                .font(.system(size: 14, weight: .bold))
                                                .frame(maxWidth: .infinity)
                                                .padding(.vertical, 10)
                                                .background(isSelected ? Color.blue : Color.white.opacity(0.06))
                                                .foregroundStyle(isSelected ? .white : .white.opacity(0.7))
                                                .cornerRadius(8)
                                        }
                                    }
                                }
                            }
                            Text("Если не выбран ни один класс, мероприятие доступно всем.")
                                .font(.system(size: 12))
                                .foregroundStyle(.white.opacity(0.4))
                        }
                    }
                    .padding(20)
                }
            }
            .navigationTitle("Новое событие")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("Отмена") { dismiss() }
                        .foregroundStyle(.white.opacity(0.7))
                }
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Создать") {
                        createEvent()
                    }
                    .font(.system(size: 16, weight: .bold))
                    .foregroundStyle(.blue)
                    .disabled(title.isEmpty || rewardString.isEmpty || isSubmitting)
                }
            }
        }
    }
    
    private func createEvent() {
        guard let reward = Int(rewardString) else {
            errorMessage = "Награда должна быть числом"
            return
        }
        
        isSubmitting = true
        errorMessage = nil
        
        // Форматируем дату в ISO8601 (обязательно для бэкенда на Go)
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        let dateString = formatter.string(from: startDate)
        
        Task {
            do {
                try await viewModel.createEvent(
                    title: title,
                    description: description,
                    reward: reward,
                    startedAt: dateString,
                    classes: Array(selectedClasses)
                )
                dismiss()
            } catch {
                self.errorMessage = "Ошибка при создании: \(error.localizedDescription)"
                self.isSubmitting = false
            }
        }
    }
}
