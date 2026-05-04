import {UserRoles, type UserType} from "../entities/user/types/user_types.ts";

interface Props {
    user: UserType,
    onClick?: () => void,
}

export function UserCard({user, onClick}: Props) {
    
    return (
        <button
            className={"class-flat-card"}
            key={user.Id}
            onClick={() => {
                if (onClick) {
                    void onClick();
                }
            }}
        >
            
            <div className={"class-flat-card__main"}>
                <div className="class-flat-card__icon">
                    {user.Name.split("")[0]}{user.LastName.split("")[0]}
                </div>
                <div className={"class-flat-card__info"}>
                    <h2 className={"class-flat-card__subtitle"}>{user.Name} {user.LastName}</h2>
                    <h2 className={"class-flat-card__subtitle"}>Рейтинг: {user.Rating}</h2>
                </div>
            </div>
            
            <div className={"class-flat-card__meta"}>
                <div className={"class-flat-card__metric"}>
                    <span className={"class-flat-card__metric-label"}>Класс</span>
                    <span className={"class-flat-card__metric-value"}>{user.Class}</span>
                </div>
                
                <div className={"class-flat-card__metric"}>
                    <span className={"class-flat-card__metric-label"}>Роль</span>
                    <span className={"class-flat-card__metric-value"}>{UserRoles[user.Role]}</span>
                </div>
            </div>
        </button>
    );
}