import { createBrowserRouter } from "react-router-dom";
import {ProfilePage} from "../../features/profile/ui/profile_page.tsx";
import {LoginPage} from "../../features/auth/ui/login_page.tsx";
import {ProtectedRoute} from "../../features/auth/ui/protected_route.tsx";
import {DashboardPage} from "../../features/dashboard/ui/pages/dashboard_page.tsx";
import {ClassDashboard} from "../../features/class_dashboard/ui/pages/class_dashboard.tsx";
import {UserPage} from "../../features/user/ui/pages/user_page.tsx";
import { ErrorBoundary } from "../error/error_boundary";
import type {ReactNode} from "react";
import {EventPage} from "../../features/events/ui/pages/event_page.tsx";
import {EventClassPlayersPage} from "../../features/events/ui/pages/event_class_players_page.tsx";

const withBoundary = (element: ReactNode, name: string) => (
    <ErrorBoundary
        fallback={
          <div className="page">
            <div className="alert alert--danger">
              Ошибка в странице {name}. Попробуйте обновить.
            </div>
          </div>
        }
    >
      {element}
    </ErrorBoundary>
);

export const router = createBrowserRouter([
  {
    path: "/login",
    element: withBoundary(<LoginPage />, "входа"),
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        path: "/",
        element: withBoundary(<DashboardPage />, "главной"),
      },
      {
        path: "/profile",
        element: withBoundary(<ProfilePage />, "профиля"),
      },
      {
        path: "/class/:id",
        element: withBoundary(<ClassDashboard />, "класса"),
      },
      {
        path: "/user/:id",
        element: withBoundary(<UserPage />, "пользователя"),
      },
      {
        path: "/event",
        element: withBoundary(<EventPage />, "мероприятия"),
      },
      {
        path: "/events/:eventId/classes/:classId/players/add",
        element: withBoundary(<EventClassPlayersPage/>, "мероприятия"),
      }
    ],
  },
]);