import { RouterProvider } from "react-router-dom";
import { router } from "./router";
import { AuthInitializer } from "../../features/auth/ui/auth_initializer";
import { ErrorBoundary } from "../../core/error/error_boundary";

export default function App() {
    return (
        <ErrorBoundary
            onError={(error, info) => {
                if (import.meta.env.DEV) {
                    console.error("App Error:", error, info);
                }
            }}
        >
            <AuthInitializer>
                <RouterProvider router={router} />
            </AuthInitializer>
        </ErrorBoundary>
    );
}