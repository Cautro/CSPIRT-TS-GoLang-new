type AuthMode = "bearer-memory" | "cookie";

function getAuthMode(): AuthMode {
    const value = import.meta.env.VITE_AUTH_MODE;

    if (value === "cookie") return "cookie";

    return "bearer-memory";
}

export class AppConfig {
    static readonly API_URL = import.meta.env.VITE_API_URL ?? "/backend";
    static readonly AUTH_MODE: AuthMode = getAuthMode();
    static readonly IS_PROD = import.meta.env.PROD;
    static readonly REQUEST_TIMEOUT_MS = 10_000;
}