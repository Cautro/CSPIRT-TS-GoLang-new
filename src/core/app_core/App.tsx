import { RouterProvider } from "react-router-dom";
import { router } from "./router";
import { AuthInitializer } from "../../features/auth/ui/auth_initializer";

export default function App() {
    return (
        <AuthInitializer>
            <RouterProvider router={router} />
        </AuthInitializer>
    );
}