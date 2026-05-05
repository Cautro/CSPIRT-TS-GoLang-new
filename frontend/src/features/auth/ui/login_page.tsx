import { type FormEvent, useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";

import {
    LOGIN_REGEX,
    SECURITY_LIMITS,
} from "../../../core/security/security_limits";

import { useAuthStore } from "../store/auth_store";

export function LoginPage() {
    const navigate = useNavigate();

    const login = useAuthStore((state) => state.login);
    const status = useAuthStore((state) => state.status);
    const error = useAuthStore((state) => state.error);

    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");

    const isLoading = status === "loading";

    if (status === "authenticated") {
        return <Navigate to="/" replace />;
    }

    const trimmedUsername = username.trim();

    const usernameError =
        username.length > 0 && !LOGIN_REGEX.test(username)
            ? "Логин может содержать только латиницу, цифры, точку, дефис и подчёркивание"
            : null;

    const passwordError =
        password.length > 0 && password.length < SECURITY_LIMITS.passwordMin
            ? `Пароль должен быть не короче ${SECURITY_LIMITS.passwordMin} символов`
            : null;

    const isSubmitDisabled =
        isLoading ||
        trimmedUsername.length < SECURITY_LIMITS.loginMin ||
        trimmedUsername.length > SECURITY_LIMITS.loginMax ||
        password.length < SECURITY_LIMITS.passwordMin ||
        password.length > SECURITY_LIMITS.passwordMax ||
        Boolean(usernameError) ||
        Boolean(passwordError);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        const success = await login({
            login: trimmedUsername,
            password,
        });

        if (success) {
            navigate("/", { replace: true });
        }
    }

    return (
        <main className="auth-page">
            <section className="auth-card">
                <div className="auth-card__header">

                    <div>
                        <h1 className="auth-card__title">Вход в систему</h1>
                    </div>
                </div>

                {error && <div className="alert alert--danger auth-card__alert">{error}</div>}

                <form className="auth-form" onSubmit={handleSubmit}>
                    <div className="field">
                        <label className="field__label" htmlFor="login">
                            Логин
                        </label>

                        <input
                            id="login"
                            className={usernameError ? "input input--error" : "input"}
                            value={username}
                            onChange={(event) => setUsername(event.target.value)}
                            placeholder="Введите ваш логин"
                            type="text"
                            autoComplete="username"
                            disabled={isLoading}
                            maxLength={SECURITY_LIMITS.loginMax}
                        />

                        <div className={usernameError ? "field__error" : "field__hint"}>
                            {usernameError}
                        </div>
                    </div>

                    <div className="field">
                        <label className="field__label" htmlFor="password">
                            Пароль
                        </label>

                        <input
                            id="password"
                            className={passwordError ? "input input--error" : "input"}
                            value={password}
                            onChange={(event) => setPassword(event.target.value)}
                            placeholder="Введите пароль"
                            type="password"
                            autoComplete="current-password"
                            disabled={isLoading}
                            minLength={SECURITY_LIMITS.passwordMin}
                            maxLength={SECURITY_LIMITS.passwordMax}
                        />

                        <div className={passwordError ? "field__error" : "field__hint"}>
                            {passwordError}
                        </div>
                    </div>

                    <button
                        className="btn btn--primary btn--lg auth-submit"
                        type="submit"
                        disabled={isSubmitDisabled}
                    >
                        {isLoading ? (
                            <span className="auth-submit__loading">
                <span className="spinner" />
                Вход...
              </span>
                        ) : (
                            "Войти"
                        )}
                    </button>
                </form>
            </section>
        </main>
    );
}