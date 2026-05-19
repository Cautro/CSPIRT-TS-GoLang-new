import {UserCard} from "../../../../shared/ui/cards/user_card.tsx";
import {useUsersStore} from "../../store/users_store.ts";
import {useNavigate} from "react-router-dom";
import {useEffect} from "react";

type props = {
    id: number,
    name: string,
}

export function UsersWidget({id, name}: props) {
    const navigate = useNavigate();
    const users = useUsersStore((state) => state.users);
    const getUsers = useUsersStore((state) => state.getUsersByClass);
    const status = useUsersStore((state) => state.status);
    const error = useUsersStore((state) => state.error);
    
    const isLoading = status === "loading";
    
    useEffect(() => {
        void getUsers(id);
    }, [getUsers, id])

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