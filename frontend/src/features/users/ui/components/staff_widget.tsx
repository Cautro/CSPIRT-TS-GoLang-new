import {useEffect} from "react";
import {useUsersStore} from "../../store/users_store.ts";
import {StaffCard} from "../../../../shared/ui/cards/staff_card.tsx";

export function StaffWidget() {
    const users = useUsersStore((state) => state.staff);
    const getStaff = useUsersStore((state) => state.getStaff);
    const status = useUsersStore((state) => state.status);
    const error = useUsersStore((state) => state.error);

    const isLoading = status === "loading";

    useEffect(() => {
        void getStaff();
    }, [getStaff])

    return (
        <>
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

            {users && !isLoading && !error && users.length > 0 ? (
                <div className="class-list">
                    {users.map((item) => (
                        <StaffCard key={item.Id} user={item} />
                    ))}
                </div>
            ) : (
                !isLoading && !error && (users?.length === 0 || !users) &&  <div className="empty-state">
                    <h2 className="empty-state__title">Персонал не найден</h2>
                    <p className="empty-state__text">
                        Не удалось найти персонал
                    </p>
                </div>
            )}
        </>
    );
} 