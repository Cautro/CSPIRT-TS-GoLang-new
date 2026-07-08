import SwiftUI

struct ScheduleView: View {
    @StateObject private var viewModel = ScheduleViewModel()
    @Environment(\.colorScheme) private var colorScheme
    
    let cardBackground = Color.white.opacity(0.06)
    
    var body: some View {
        ZStack {
            backgroundLayer
                .ignoresSafeArea()
            
            ScrollView(showsIndicators: false) {
                VStack(spacing: 24) {
                    
                    headerCard
                    
                    dayNavigation
                    
                    VStack(spacing: 12) {
                        if viewModel.isLoading {
                            ProgressView()
                                .tint(.white)
                                .padding(.top, 40)
                        } else if viewModel.filteredCurrentSchedules.isEmpty {
                            Text("Нет уроков в этот день")
                                .foregroundColor(.gray)
                                .padding(.top, 40)
                        } else {
                            ForEach(viewModel.filteredCurrentSchedules) { lesson in
                                LessonRowView(lesson: lesson, cardBackground: cardBackground)
                            }
                        }
                    }
                }
                .padding(.horizontal, 35)
            }
        }
        .onAppear {
            Task {
                await viewModel.fetchSchedule()
            }
        }
    }
    
    // MARK: - Subviews
    
    private var headerCard: some View {
        HStack(spacing: 16) {
            Image(systemName: "calendar.badge.clock")
                .font(.system(size: 32, weight: .regular))
                .foregroundColor(.white)
                .frame(width: 50, height: 50)
            
            VStack(alignment: .leading, spacing: 4) {
                Text("Расписание уроков")
                    .font(.title3)
                    .fontWeight(.bold)
                    .foregroundColor(.white)
                
                Text("Текущий график на неделю")
                    .font(.subheadline)
                    .foregroundColor(.gray)
            }
            Spacer()
        }
        .padding(20)
        .background(cardBackground)
        .cornerRadius(24)
        .padding(.horizontal)
    }
    
    private var dayNavigation: some View {
        HStack {
            // Кнопка "Предыдущий день"
            Button(action: {
                withAnimation(.easeInOut(duration: 0.2)) {
                    let allDays = WeekDay.allCases
                    if let currentIndex = allDays.firstIndex(of: viewModel.selectedDay) {
                        // Вычисляем предыдущий индекс с зацикливанием
                        let prevIndex = (currentIndex - 1 + allDays.count) % allDays.count
                        viewModel.selectedDay = allDays[prevIndex]
                    }
                }
            }) {
                Image(systemName: "chevron.left")
                    .font(.system(size: 20, weight: .semibold))
                    .foregroundColor(.white)
                    .frame(width: 48, height: 48)
                    .background(cardBackground)
                    .cornerRadius(16)
            }
            
            Spacer()
            
            // Название выбранного дня
            Text(viewModel.selectedDay.displayName)
                .font(.title2)
                .fontWeight(.bold)
                .foregroundColor(.white)
                // Анимация смены текста
                .id(viewModel.selectedDay.displayName)
                .transition(.opacity)
            
            Spacer()
            
            // Кнопка "Следующий день"
            Button(action: {
                withAnimation(.easeInOut(duration: 0.2)) {
                    let allDays = WeekDay.allCases
                    if let currentIndex = allDays.firstIndex(of: viewModel.selectedDay) {
                        // Вычисляем следующий индекс с зацикливанием
                        let nextIndex = (currentIndex + 1) % allDays.count
                        viewModel.selectedDay = allDays[nextIndex]
                    }
                }
            }) {
                Image(systemName: "chevron.right")
                    .font(.system(size: 20, weight: .semibold))
                    .foregroundColor(.white)
                    .frame(width: 48, height: 48)
                    .background(cardBackground)
                    .cornerRadius(16)
            }
        }
        .padding(.horizontal)
    }
}

private extension ScheduleView {
    
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

// MARK: - Компонент Карточки Урока

struct LessonRowView: View {
    let lesson: ScheduleModel
    let cardBackground: Color
    
    var body: some View {
//        GlassCard{
//            Text("\(lesson.lessonNumber)")
////                    .font(.title2)
////                    .fontWeight(.bold)
////                    .foregroundColor(.blue)
////                    .frame(width: 48, height: 48)
////                .background(Color.blue.opacity(0.1))
////                                .overlay(
////                                    Square().stroke(Color.blue.opacity(0.3), lineWidth: 1)
////                                )
////                    .clipShape(Circle())
//        }

        HStack(spacing: 16) {
            
            // Информация об уроке
            VStack(alignment: .leading, spacing: 6) {
                Text("\(lesson.startTime) – \(lesson.endTime)")
                    .font(.caption)
                    .foregroundColor(.secondary)
                
                Text(lesson.subject)
                    .font(.title3)
                    .fontWeight(.bold)
                    .foregroundColor(.white)
                
                HStack(spacing: 16) {
                    HStack(spacing: 4) {
                        Image(systemName: "person.fill")
                        
                        if let nameData = lesson.teacher.fullName.first {
                            Text("\(nameData.LastName ?? "") \(nameData.Name ?? "") \(nameData.MiddleName ?? "")")
                        } else {
                            Text("Нет данных об учителе")
                        }
                    }
                    
                    HStack(spacing: 4) {
                        Image(systemName: "door.left.hand.closed")
                        Text("Каб. \(lesson.room)")
                    }
                }
                .font(.caption)
                .foregroundColor(Color.primary)
            }
            
            Spacer()
        }
        .padding(20)
        .background(cardBackground)
        .cornerRadius(24)
    }
}

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
