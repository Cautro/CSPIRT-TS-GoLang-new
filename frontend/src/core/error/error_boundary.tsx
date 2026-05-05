import { Component, type ReactNode, type ErrorInfo } from "react";

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
    onError?: (error: Error, info: ErrorInfo) => void;
}

interface State {
    hasError: boolean;
    error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, info: ErrorInfo): void {
        this.props.onError?.(error, info);
    }

    render(): ReactNode {
        if (this.state.hasError) {
            return this.props.fallback ?? (
                <div className="empty-state">
                    <h2 className="empty-state__title">Что-то пошло не так</h2>
                    <p className="empty-state__text">
                        Произошла ошибка. Попробуйте обновить страницу.
                    </p>
                    <button
                        className="btn btn--primary"
                        onClick={() => this.setState({ hasError: false, error: null })}
                    >
                        Попробовать снова
                    </button>
                </div>
            );
        }
        return this.props.children;
    }
}