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
import type {NoteType} from "../../../shared/entities/notes/notes_types.ts";

interface UserDashboardPageProps {
    users: UserType[];
    notes: NoteType[];
    status: DashboardStatus;
    error: string | null;
    role: string;
    getUsers: () => Promise<void>;
}

type selectedDashboard = | "users" | "notes";

export function UserDashboardPage({
                                      users,
                                      notes,
                                      status,
                                      error,
                                      role,
                                      getUsers,
                                  }: UserDashboardPageProps) {
    const navigate = useNavigate();

    const [showActions, setShowActions] = useState(false);
    const [selectedUser, setSelectedUser] = useState<UserType | null>(null);
    const [selectedDashboard, setSelectedDashboard] = useState<selectedDashboard>("users");

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
                        
                        {(role === "Admin" || role === "Owner") && (
                            <Box>
                                <Button
                                    variant="contained"
                                    onClick={() => setSelectedDashboard("users")}
                                    disabled={selectedDashboard === "users"}
                                >
                                    Пользователи
                                </Button>
                                
                                <Button
                                    variant="contained"
                                    onClick={() => setSelectedDashboard("notes")}
                                    disabled={selectedDashboard === "notes"}
                                >
                                    Заметки
                                </Button>
                            </Box>
                        )}
                        
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

                {selectedDashboard === "notes" && (
                    notes.length > 0 ? (
                        <Box sx={{ display: "grid", gap: 2 }}>
                            {notes.map((note) => (
                                <Box key={note.ID}>
                                    <Typography>Контент: {note.Content}</Typography>
                                </Box>
                            ))}
                        </Box>
                    ) : (
                        !isLoading && <Alert severity="info">Заметки не найдены</Alert>
                    )
                )}

                {selectedDashboard === "users" && (
                    users.length > 0 ? (
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
                        !isLoading && <Alert severity="info">Пользователи не найдены</Alert>
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