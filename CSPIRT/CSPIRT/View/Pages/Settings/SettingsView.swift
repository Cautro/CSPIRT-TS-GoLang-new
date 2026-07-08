import SwiftUI

struct SettingsView: View {

    @EnvironmentObject var settings: AppSettings
    @EnvironmentObject var session: SessionManager
    //@ObservableObject var cache: AppCacheStore
    
    @State private var isShowLogoutModal: Bool = false
    @State private var isShowClearCacheModal: Bool = false
    @State private var isCacheAvailable: Bool = false
    
    @State private var cacheSize: Double = 0.0

    var body: some View {
        ZStack {
            Color.black.ignoresSafeArea()

            Form {
                Section("Общее") {

                    Picker("Тема", selection: $settings.themeMode) {
                        Text("Система").tag("system")
                        Text("Тёмная").tag("dark")
                        Text("Светлая").tag("light")
                    }

//                    Toggle("Уведомления", isOn: $settings.notificationsEnabled)
//
//                    Toggle("Тактильная отдача", isOn: $settings.hapticsEnabled)
                }
                
                Section("Кеш") {
                    Text("Размер кеша: \(String(format: "%.2f", cacheSize)) МБ")
                    Picker("Длительность кеша", selection: $settings.longCache) {
                        Text("1 день").tag("1 day")
                        Text("2 дня").tag("2 days")
                        Text("4 дня").tag("4 days")
                        Text("7 дней").tag("7 days")
                        Text("14 дней").tag("14 days")
                        Text("30 дней").tag("30 days")
                        Text("120 дней").tag("120 days")
                    }
                    
                    Button {
                        self.isShowClearCacheModal = true
                    } label: {
                        Text("Отчистить весь кеш")
                    }
                    .disabled(!isCacheAvailable)
                }

                Section("Аккаунт") {
                    Button(role: .destructive) {
                        self.isShowLogoutModal = true
                    } label: {
                        Text("Выйти из аккаунта")
                    }
                }
                
                Section("Миниигры") {
                    NavigationStack {
                        NavigationLink(destination: BarankaGameView()) {
                            Button() {
                            } label: {
                                Text("Баранкакетч")
                            }
                        }
                    }
                }

                Section("Информация") {
                    Text("Версия: 1.0.0 (Бета-версия)")
                        .foregroundStyle(.secondary)
                }
            }
        }
        .navigationTitle("Настройки")
        
        .task {
            let size = await AppCacheStore.shared.getCacheSizeInMB()
            self.cacheSize = size
            isCacheAvailable = size > 0.5
        }
        
        .alert("Вы уверены, что хотите выйти?", isPresented: $isShowLogoutModal) {
            Button("Выйти", role: .destructive) {
                session.logout()
            }
            Button("Отмена", role: .cancel) {}
            
        } message: {
            Text("Для повторного входа вам потребуется ввести свои данные.")
        }
        
        .alert("Вы уверены, что хотите отчистить весь кеш?", isPresented: $isShowClearCacheModal) {
            Button("Отчистить", role: .destructive) {
                Task {
                    await AppCacheStore.shared.clearAll()
                }
            }
            Button("Отмена", role: .cancel) {}
            
        } message: {
            Text("Приложение может работать медленнее если вы полностью его отчистите")
        }
    }
}
