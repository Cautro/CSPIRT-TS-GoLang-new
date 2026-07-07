import {StaffCard} from "../../../../shared/ui/cards/staff_card.tsx";
import {useStaff} from "../../hooks/use_staff.ts";
import type {UserType} from "../../../../shared/entities/user/types/user_types.ts";

export function StaffWidget() {
    const {data, error, isLoading} = useStaff();
    const users = (data as UserType[]) || [];
    
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
                <div className="alert alert--danger mb-4">{error.message}</div>
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