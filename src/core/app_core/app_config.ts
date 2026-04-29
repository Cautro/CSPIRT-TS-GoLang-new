export class AppConfig {
    static readonly API_URL = import.meta.env.VITE_API_URL ?? "";
    static readonly IS_PROD = import.meta.env.PROD;
    static readonly REQUEST_TIMEOUT_MS = 10_000;
}