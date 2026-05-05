export const SECURITY_LIMITS = {
    loginMin: 3,
    loginMax: 64,
    passwordMin: 6,
    passwordMax: 128,
    ratingReasonMin: 5,
    ratingReasonMax: 500,
    complaintMax: 1000,
    noteMax: 1000,
    ratingDeltaMin: -5000,
    ratingDeltaMax: 5000,
} as const;

export const LOGIN_REGEX = /^[a-zA-Z0-9._-]+$/;

export function normalizeText(value: string): string {
    return value.trim().replace(/\s+/g, " ");
}

export function truncateText(value: string, max = 300): string {
    if (value.length <= max) return value;
    return `${value.slice(0, max)}...`;
}