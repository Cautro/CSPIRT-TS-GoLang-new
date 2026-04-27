export class ApiResponse<T> {
    data: T | undefined;
    status_code: number;
    
    constructor(data: T, status_code: number) {
        this.data = data;
        this.status_code = status_code;
    }
    
    static async fromResponse(res: Response) {
        return new ApiResponse(await res.json(), res.status);
    }
    
    checkStatus() {
        return this.status_code >= 200 && this.status_code <= 299;
    }
}