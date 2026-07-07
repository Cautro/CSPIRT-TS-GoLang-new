import {UserCard} from "../../../../shared/ui/cards/user_card.tsx";
import {useNavigate} from "react-router-dom";
import {UseUsersByClass} from "../../hooks/use_users_by_class.ts";
import type {UserType} from "../../../../shared/entities/user/types/user_types.ts";

type props = {
    id: number,
    name: string,
}

export function UsersWidget({id, name}: props) {
    const navigate = useNavigate();
    const {data, error, isLoading} = UseUsersByClass(id);
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
            
            {(users !== null && users.length > 0) ? (
                <div className={"class-list"}>
                    {users.map((user) => (
                        <UserCard user={user} key={user.Id} onClick={() => navigate(`/user/${user.Id}`)}/>
                    ))}
                </div>
            ) : (
                !isLoading && <div className="empty-state">
                    <h2 className="empty-state__title">Ученики не найдены</h2>
                    <p className="empty-state__text">
                        Не удалось найти учеников по {name} классу
                    </p>
                </div>
            )}
        </>
    );
} 