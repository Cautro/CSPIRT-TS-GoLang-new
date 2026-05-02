import { createBrowserRouter } from "react-router-dom";
import {ProfilePage} from "../../features/profile/ui/profile_page.tsx";
import {LoginPage} from "../../features/auth/ui/login_page.tsx";
import {ProtectedRoute} from "../../features/auth/ui/protected_route.tsx";
import {DashboardPage} from "../../features/dashboard/ui/pages/dashboard_page.tsx";
import {ClassDashboard} from "../../features/class_dashboard/ui/pages/class_dashboard.tsx";
import {UserPage} from "../../features/user/ui/pages/user_page.tsx";

export const router = createBrowserRouter([
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        path: "/",
        element: <DashboardPage />,
      },
      {
        path: "/profile",
        element: <ProfilePage />,
      },
      {
        path: "/classDashboard",
        element: <ClassDashboard />
      },
      {
        path: "/user/:id",
        element: <UserPage />
      }
    ],
  },
]);