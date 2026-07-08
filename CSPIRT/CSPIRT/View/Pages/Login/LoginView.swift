import SwiftUI

struct LoginRequest: Codable {
    let login: String
    let password: String
    
    enum CodingKeys: String, CodingKey {
        case login = "Login"
        case password = "Password"
    }
}

private enum Field: Hashable {
    case login
    case password
}

struct LoginView: View {
    @StateObject private var viewModel = LoginViewModel()
    @EnvironmentObject private var sessionManager: SessionManager
    
    @State private var isPasswordVisible = false
    @State private var rememberMe = false
    
    @FocusState private var focusedField: Field?
    
    var body: some View {
        ZStack {
            // Задний фон на весь экран
            backgroundLayer
                .ignoresSafeArea()

            // Основной контейнер с адаптивным поведением
            GeometryReader { proxy in
                ScrollView(.vertical, showsIndicators: false) {
                    
                    headerSection
                        .padding(.top, proxy.safeAreaInsets.top)
                        .padding(.bottom, 24)
                        //.padding(.bottom, proxy.safeAreaInsets.bottom)
                    VStack(spacing: 0) {

                        formSection
                            .padding(.top, 24)

                        Spacer(minLength: 32)

                        if let error = viewModel.errMsg, !error.isEmpty {
                            errorView(message: error)
                                .padding(.horizontal, 24)
                                .padding(.bottom, 16)
                        }

                        confirmButton
                            .padding(.horizontal, 24)
                            .padding(.bottom, max(16, proxy.safeAreaInsets.bottom + 12))
                    }
                    .frame(width: proxy.size.width)
                    .background(Color.gray.opacity(0.125)).cornerRadius(35)
                    
                }
                .scrollDismissesKeyboard(.interactively)
            }
            
        }
        .onTapGesture {
            focusedField = nil
        }
        .animation(.snappy, value: viewModel.errMsg)
        .animation(.snappy, value: viewModel.isLoading)
    }
}

private extension LoginView {

    // Размытый темный фон
    var backgroundLayer: some View {
        ZStack {
            Image("school_background")
                .resizable()
                //.scaledToFill()
                .blur(radius: 10)

            LinearGradient(
                colors: [
                    Color.black.opacity(0.4),
                    Color.black.opacity(0.8),
                    Color.black.opacity(0.95)
                ],
                startPoint: .top,
                endPoint: .bottom
            )
        }
    }

    // Логотип и заголовки
    var headerSection: some View {
        VStack(spacing: 16) {
            ZStack {
                Circle()
                    .fill(.white.opacity(0.12))
                    .frame(width: 96, height: 96)
                    .overlay(Circle().stroke(Color.white.opacity(0.2), lineWidth: 1))

                Image(systemName: "book.closed.fill")
                    .font(.system(size: 38, weight: .bold))
                    .foregroundStyle(.white)

                Image(systemName: "chart.bar.fill")
                    .font(.system(size: 14, weight: .bold))
                    .foregroundStyle(Color(red: 0.0, green: 0.55, blue: 1.0))
                    .offset(x: 4, y: 0)
            }
            .shadow(color: .black.opacity(0.3), radius: 12)

            VStack(spacing: 6) {
                Text("Социальный рейтинг")
                    .font(.system(size: 26, weight: .bold, design: .rounded))
                    .foregroundStyle(.white)

                Text("МАОУ СОШ №16Ф")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(.white.opacity(0.6))
            }
        }
    }

    // Поля ввода
    var formSection: some View {
        VStack(spacing: 24) {
            VStack(spacing: 6) {
                Text("Вход в систему")
                    .font(.system(size: 22, weight: .bold, design: .rounded))
                    .foregroundStyle(.white)

                Text("Используйте свой логин и пароль")
                    .font(.system(size: 13, weight: .regular))
                    .foregroundStyle(.white.opacity(0.45))
            }

            // Вертикальный блок инпутов
            VStack(spacing: 14) {
                // ПОЛЕ ЛОГИНА
                HStack(spacing: 14) {
                    Image(systemName: "person.fill")
                        .font(.system(size: 16))
                        .foregroundStyle(.white.opacity(0.5))
                        .frame(width: 20)

                    TextField("", text: $viewModel.Login, prompt: placeholder("Логин*"))
                        .font(.system(size: 16, weight: .medium))
                        .textInputAutocapitalization(.never)
                        .disableAutocorrection(true)
                        .foregroundStyle(.white)
                        .focused($focusedField, equals: .login)
                        .submitLabel(.next)
                        .onSubmit { focusedField = .password }
                }
                .modifier(MobileFieldModifier())

                // ПОЛЕ ПАРОЛЯ
                HStack(spacing: 14) {
                    Image(systemName: "lock.fill")
                        .font(.system(size: 16))
                        .foregroundStyle(.white.opacity(0.5))
                        .frame(width: 20)

                    Group {
                        if isPasswordVisible {
                            TextField("", text: $viewModel.Pass, prompt: placeholder("Пароль*"))
                        } else {
                            SecureField("", text: $viewModel.Pass, prompt: placeholder("Пароль*"))
                        }
                    }
                    .font(.system(size: 16, weight: .medium))
                    .textInputAutocapitalization(.never)
                    .disableAutocorrection(true)
                    .foregroundStyle(.white)
                    .focused($focusedField, equals: .password)
                    .submitLabel(.go)
                    .onSubmit { loginAction() }

                    Button(action: { isPasswordVisible.toggle() }) {
                        Image(systemName: isPasswordVisible ? "eye" : "eye.slash")
                            .font(.system(size: 16))
                            .foregroundStyle(.white.opacity(0.5))
                            .frame(width: 24, height: 24)
                    }
                }
                .modifier(MobileFieldModifier())
            }
            .padding(.horizontal, 24)

            // Чекбокс Запомнить меня
            HStack {
                Toggle(isOn: $rememberMe) {
                    Text("Запомнить меня")
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(.white.opacity(0.7))
                }
                .toggleStyle(MobileCheckboxToggleStyle())
                Spacer()
            }
            .padding(.horizontal, 28)
        }
    }

    // Кнопка подтверждения
    var confirmButton: some View {
        Button(action: loginAction) {
            ZStack {
                RoundedRectangle(cornerRadius: 14, style: .continuous)
                    .fill(Color(red: 0.0, green: 0.6, blue: 1.0))

                if viewModel.isLoading {
                    ProgressView()
                        .tint(.white)
                } else {
                    Text("Подтвердить")
                        .font(.system(size: 16, weight: .bold, design: .rounded))
                        .foregroundStyle(.white)
                }
            }
            .frame(height: 54)
            .shadow(color: Color(red: 0.0, green: 0.6, blue: 1.0).opacity(0.25), radius: 8, x: 0, y: 4)
        }
        .disabled(viewModel.isLoading)
        .opacity(viewModel.isLoading ? 0.8 : 1.0)
    }
}

private extension LoginView {
    func loginAction() {
        focusedField = nil
        Task {
            await viewModel.performLogin()
        }
    }

    func placeholder(_ text: String) -> Text {
        Text(text).foregroundStyle(.white.opacity(0.35))
    }

    func errorView(message: String) -> some View {
        HStack(spacing: 10) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 14))
                .foregroundColor(.red)
            Text(message)
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(.red.opacity(0.9))
        }
        .padding(.horizontal, 16)
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

// Кастомный модификатор для полей
struct MobileFieldModifier: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding(.horizontal, 16)
            .frame(height: 54)
            .background {
                RoundedRectangle(cornerRadius: 12, style: .continuous)
                    .fill(Color.white.opacity(0.06))
                    .overlay(
                        RoundedRectangle(cornerRadius: 12, style: .continuous)
                            .stroke(Color.white.opacity(0.12), lineWidth: 1)
                    )
            }
    }
}

// Удобный мобильный стиль чекбокса
struct MobileCheckboxToggleStyle: ToggleStyle {
    func makeBody(configuration: Configuration) -> some View {
        Button(action: {
            withAnimation(.snappy(duration: 0.15)) {
                configuration.isOn.toggle()
            }
        }) {
            HStack(spacing: 10) {
                Image(systemName: configuration.isOn ? "checkmark.square.fill" : "square")
                    .font(.system(size: 20))
                    .foregroundStyle(configuration.isOn ? Color(red: 0.0, green: 0.6, blue: 1.0) : .white.opacity(0.4))

                configuration.label
            }
        }
        .buttonStyle(.plain)
    }
}
