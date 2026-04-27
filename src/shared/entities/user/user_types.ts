
export interface FullName {
    name: string;
    lastname: string;
}

export interface UserType {
    Id: number;
    Name: string;
    LastName: string;
    FullName: FullName[];
    Login: string;
    Rating: number;
    Role: string;
    Class: string;
    Notes: unknown[];
    Complaints: unknown[];
}

export const UserRoles = {
    "Admin": "Администратор",
    "User": "Пользователь",
    "Owner": "Owner",
    "Helper": "Староста"
}