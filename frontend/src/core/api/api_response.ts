export class ApiResponse<T> {
    data: T | undefined;
    status_code: number;

    constructor(data: T | undefined, status_code: number) {
        this.data = data;
        this.status_code = status_code;
    }

    static async fromResponse<T>(res: Response): Promise<ApiResponse<T>> {
        const contentType = res.headers.get("content-type") ?? "";
        const text = await res.text();

        if (!text) {
            return new ApiResponse<T>(undefined, res.status);
        }

        if (!contentType.includes("application/json")) {
            return new ApiResponse<T>(undefined, res.status);
        }

        try {
            return new ApiResponse<T>(JSON.parse(text) as T, res.status);
        } catch {
            return new ApiResponse<T>(undefined, res.status);
        }
    }

    checkStatus() {
        return this.status_code >= 200 && this.status_code <= 299;
    }
}