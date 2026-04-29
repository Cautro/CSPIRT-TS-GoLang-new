import { useState } from "react";
import {
    Alert,
    Box,
    Button,
    Container,
    LinearProgress,
    Typography,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import { UserActions } from "./user_actions.tsx";
import type { UserType } from "../../../shared/entities/user/user_types.ts";
import type { DashboardStatus } from "../store/dashboard_store.ts";

interface UserDashboardPageProps {
    users: UserType[];
    status: DashboardStatus;
    error: string | null;
    role: string;
    getUsers: () => Promise<void>;
}

export function UserDashboardPage({
                                      users,
                                      status,
                                      error,
                                      role,
                                      getUsers,
                                  }: UserDashboardPageProps) {
    const navigate = useNavigate();

    const [showActions, setShowActions] = useState(false);
    const [selectedUser, setSelectedUser] = useState<UserType | null>(null);

    const isLoading = status === "loading";

    return (
        <Container maxWidth="lg">
            <Box sx={{ py: 4 }}>
                <Box
                    sx={{
                        mb: 3,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "space-between",
                        gap: 2,
                    }}
                >
                    <Typography variant="h4" sx={{ fontWeight: 800 }}>
                        Дашборд рейтинга
                    </Typography>

                    <Box sx={{ display: "flex", gap: 2 }}>
                        <Button
                            variant="contained"
                            onClick={() => void getUsers()}
                            disabled={isLoading}
                        >
                            Обновить
                        </Button>

                        <Button
                            variant="contained"
                            onClick={() => navigate("/profile")}
                            disabled={isLoading}
                        >
                            Профиль
                        </Button>
                    </Box>
                </Box>

                {isLoading && <LinearProgress sx={{ mb: 2 }} />}

                {error && (
                    <Alert severity="error" sx={{ mb: 2 }}>
                        {error}
                    </Alert>
                )}

                {users.length > 0 ? (
                    <Box
                        sx={{
                            display: "grid",
                            gridTemplateColumns: {
                                xs: "1fr",
                                md: "1fr 1fr",
                            },
                            gap: 2,
                        }}
                    >
                        {users.map((user) => (
                            <Box
                                key={user.Id}
                                sx={{
                                    p: 2,
                                    border: "1px solid",
                                    borderColor: "divider",
                                    borderRadius: 3,
                                    cursor: "pointer",
                                    transition: "0.15s",
                                    "&:hover": {
                                        borderColor: "primary.main",
                                    },
                                }}
                                onClick={() => {
                                    setSelectedUser(user);
                                    setShowActions(true);
                                }}
                            >
                                <Typography variant="h6" sx={{ fontWeight: 700 }}>
                                    {user.Name} {user.LastName}
                                </Typography>

                                <Typography color="text.secondary">
                                    @{user.Login}
                                </Typography>

                                <Typography>Класс: {user.Class}</Typography>

                                <Typography>Роль: {user.Role}</Typography>

                                <Typography sx={{ fontWeight: 700 }}>
                                    Рейтинг: {user.Rating}
                                </Typography>
                            </Box>
                        ))}
                    </Box>
                ) : (
                    !isLoading && (
                        <Alert severity="info">
                            Пользователи не найдены
                        </Alert>
                    )
                )}
            </Box>

            {selectedUser && showActions && (
                <Box sx={{ mb: 4 }}>
                    <UserActions user={selectedUser} role={role} />
                </Box>
            )}
        </Container>
    );
}