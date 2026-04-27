import { useEffect } from "react";
import { Alert, Box, Button, Container, LinearProgress, Typography } from "@mui/material";

import { useDashboardStore } from "../store/dashboard_store.ts";

export function UserDashboardPage() {
    const users = useDashboardStore((state) => state.users);
    const status = useDashboardStore((state) => state.status);
    const error = useDashboardStore((state) => state.error);
    const getUsers = useDashboardStore((state) => state.getUsers);

    const isLoading = status === "loading";

    useEffect(() => {
        void getUsers();
    }, [getUsers]);

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

                    <Button
                        variant="contained"
                        onClick={() => void getUsers()}
                        disabled={isLoading}
                    >
                        Обновить
                    </Button>
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
                                }}
                            >
                                <Typography variant="h6" sx={{ fontWeight: 700 }}>
                                    {user.Name} {user.LastName}
                                </Typography>

                                <Typography color="text.secondary">
                                    @{user.Login}
                                </Typography>

                                <Typography>
                                    Класс: {user.Class}
                                </Typography>

                                <Typography>
                                    Роль: {user.Role}
                                </Typography>

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
        </Container>
    );
}