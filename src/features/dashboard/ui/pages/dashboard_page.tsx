import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

import { useAuthStore } from "../../../auth/store/auth_store";
import { useDashboardStore } from "../../store/dashboard_store";
import { ClassCard } from "../../../../shared/ui/class_card.tsx";

export function DashboardPage() {
    const navigate = useNavigate();

    const role = useAuthStore((state) => state.user?.User.Role);
    const status = useDashboardStore((state) => state.status);
    const error = useDashboardStore((state) => state.error);
    const classes = useDashboardStore((state) => state.classes);
    const getClasses = useDashboardStore((state) => state.getClasses);

    const isLoading = status === "loading";

    useEffect(() => {
        void getClasses();
    }, [getClasses]);

    if (!role) {
        return null;
    }

    return (
        <main className="main">
            <section className="page">
                <div className="page__head">
                    <div>
                        <h1 className="page__title">Список классов</h1>
                        <p className="page__description">
                            Просмотр классов, классных руководителей, количества учеников и общего рейтинга.
                        </p>
                    </div>

                    <div className="btn-group">
                        <button
                            className="btn btn--secondary"
                            type="button"
                            onClick={() => void getClasses()}
                            disabled={isLoading}
                        >
                            Обновить
                        </button>

                        <button
                            className="btn btn--primary"
                            type="button"
                            onClick={() => navigate("/profile")}
                            disabled={isLoading}
                        >
                            Профиль
                        </button>
                    </div>
                </div>

                {isLoading && (
                    <div className="grid grid--3">
                        <div className="skeleton" style={{ height: 160 }} />
                        <div className="skeleton" style={{ height: 160 }} />
                        <div className="skeleton" style={{ height: 160 }} />
                    </div>
                )}

                {error && !isLoading && (
                    <div className="alert alert--danger mb-4">{error}</div>
                )}

                {!isLoading && !error && classes.length > 0 && (
                    <div className="class-list">
                        {classes.map((item) => (
                            <ClassCard key={item.Id} item={item}/>
                        ))}
                    </div>
                )}

                {!isLoading && !error && classes.length === 0 && (
                    <div className="empty-state">
                        <h2 className="empty-state__title">Классы не найдены</h2>
                        <p className="empty-state__text">
                            В системе пока нет доступных классов или у вашей роли нет прав на их просмотр.
                        </p>
                    </div>
                )}
            </section>
        </main>
    );
}