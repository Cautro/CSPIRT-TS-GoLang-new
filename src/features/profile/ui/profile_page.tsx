import {
    Alert,
    Avatar,
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    Container,
    Divider,
    LinearProgress,
    Paper,
    Typography,
} from "@mui/material";

import { useAuthStore } from "../../auth/store/auth_store";
import type { ReactNode } from "react";
import {UserRoles} from "../../../shared/entities/user/user_types.ts";
import {useNavigate} from "react-router-dom";
import { truncateText } from "../../../core/security/security_limits.ts";

export function ProfilePage() {
    const navigate = useNavigate();
    
    const profile = useAuthStore((state) => state.user);
    const getProfile = useAuthStore((state) => state.checkAuth);
    const logout = useAuthStore((state) => state.logout);
    const status = useAuthStore((state) => state.status);
    const error = useAuthStore((state) => state.error);

    const isLoading = status === "loading";

    function safeUnknownToText(value: unknown): string {
        if (typeof value === "string") {
            return truncateText(value, 500);
        }

        if (
            typeof value === "object" &&
            value !== null &&
            "Text" in value &&
            typeof (value as { Text?: unknown }).Text === "string"
        ) {
            return truncateText((value as { Text: string }).Text, 500);
        }

        return "Скрыто: неизвестный формат данных";
    }

    if (!profile) {
        return (
            <Container maxWidth="lg">
                <Box sx={{ py: 4 }}>
                    {isLoading && <LinearProgress />}
                    {error ? (
                        <Alert severity="error">{error}</Alert>
                    ) : (
                        <Alert severity="info">Профиль не загружен</Alert>
                    )}
                </Box>
            </Container>
        );
    }

    const notes = profile.Notes ?? [];
    const complaints = profile.Complaints ?? [];

    const fullName = `${profile.Name ?? ""} ${profile.LastName ?? ""}`.trim();
    const initials = `${profile.Name?.[0] ?? ""}${profile.LastName?.[0] ?? ""}`;
    const ratingPercent =  (profile.Rating / 5000) * 100

    return (
        <Container maxWidth="lg">
            <Box sx={{ py: 4 }}>
                <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
                    <Paper
                        elevation={0}
                        sx={{
                            p: { xs: 2, sm: 3 },
                            borderRadius: 4,
                            border: "1px solid",
                            borderColor: "divider",
                        }}
                    >
                        <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
                            <Box
                                sx={{
                                    display: "flex",
                                    flexDirection: { xs: "column", sm: "row" },
                                    justifyContent: "space-between",
                                    alignItems: { xs: "flex-start", sm: "center" },
                                    gap: 2,
                                }}
                            >
                                <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
                                    <Avatar
                                        sx={{
                                            width: 80,
                                            height: 80,
                                            bgcolor: "primary.main",
                                            fontSize: 28,
                                            fontWeight: 700,
                                        }}
                                    >
                                        {initials}
                                    </Avatar>

                                    <Box>
                                        <Typography variant="h4" sx={{ fontWeight: 800 }}>
                                            {fullName}
                                        </Typography>

                                        <Box
                                            sx={{
                                                mt: 1,
                                                display: "flex",
                                                alignItems: "center",
                                                flexWrap: "wrap",
                                                gap: 1,
                                            }}
                                        >
                                            <Chip label={profile.Role} color="primary" size="small" />
                                            <Chip label={`Класс ${profile.Class}`} variant="outlined" size="small" />

                                            <Typography variant="body2" sx={{ color: "text.secondary" }}>
                                                @{profile.Login}
                                            </Typography>
                                        </Box>
                                    </Box>
                                </Box>

                                <Box sx={{ display: "flex", gap: 1 }}>
                                    <Button variant="outlined" onClick={() => navigate("/", {replace: true})}>
                                        Главная
                                    </Button>
                                    
                                    <Button variant="outlined" onClick={() => void getProfile()}>
                                        Обновить
                                    </Button>

                                    <Button variant="contained" color="error" onClick={() => void logout()}>
                                        Выйти
                                    </Button>
                                </Box>
                            </Box>

                            {isLoading && <LinearProgress />}

                            {error && <Alert severity="error">{error}</Alert>}
                        </Box>
                    </Paper>

                    <Box
                        sx={{
                            display: "grid",
                            gridTemplateColumns: { xs: "1fr", md: "2fr 1fr" },
                            gap: 3,
                        }}
                    >
                        <Card variant="outlined" sx={{ borderRadius: 4 }}>
                            <CardContent>
                                <Typography variant="h6" sx={{ fontWeight: 700, mb: 2 }}>
                                    Основная информация
                                </Typography>

                                <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
                                    <InfoRow label="ID пользователя" value={profile.Id} />
                                    <InfoRow label="Имя" value={profile.Name} />
                                    <InfoRow label="Фамилия" value={profile.LastName} />
                                    <InfoRow label="Полное имя" value={fullName} />
                                    <InfoRow label="Логин" value={profile.Login} />
                                    <InfoRow label="Класс" value={profile.Class} />
                                    <InfoRow label="Роль" value={UserRoles[profile.Role]} />
                                </Box>
                            </CardContent>
                        </Card>

                        <Card variant="outlined" sx={{ borderRadius: 4 }}>
                            <CardContent>
                                <Typography variant="h6" sx={{ fontWeight: 700, mb: 2 }}>
                                    Рейтинг
                                </Typography>

                                <Typography variant="h3" sx={{ fontWeight: 800 }}>
                                    {profile.Rating}
                                </Typography>

                                <Typography variant="body2" sx={{ color: "text.secondary", mb: 1.5 }}>
                                    Текущий рейтинг пользователя
                                </Typography>

                                <LinearProgress
                                    variant="determinate"
                                    value={ratingPercent}
                                    sx={{
                                        height: 10,
                                        borderRadius: 999,
                                    }}
                                />
                            </CardContent>
                        </Card>
                    </Box>

                    <Box
                        sx={{
                            display: "grid",
                            gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
                            gap: 3,
                        }}
                    >
                        <Card variant="outlined" sx={{ borderRadius: 4 }}>
                            <CardContent>
                                <Box
                                    sx={{
                                        mb: 2,
                                        display: "flex",
                                        alignItems: "center",
                                        justifyContent: "space-between",
                                        gap: 1,
                                    }}
                                >
                                    <Typography variant="h6" sx={{ fontWeight: 700 }}>
                                        Заметки
                                    </Typography>

                                    <Chip label={notes.length} size="small" variant="outlined" />
                                </Box>

                                <Divider sx={{ mb: 2 }} />

                                {notes.length > 0 ? (
                                    <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
                                        {notes.map((note, index) => (
                                            <Alert key={index} severity="info">
                                                {safeUnknownToText(note)}
                                            </Alert>
                                        ))}
                                    </Box>
                                ) : (
                                    <Typography variant="body2" sx={{ color: "text.secondary" }}>
                                        Заметок нет
                                    </Typography>
                                )}
                            </CardContent>
                        </Card>

                        <Card variant="outlined" sx={{ borderRadius: 4 }}>
                            <CardContent>
                                <Box
                                    sx={{
                                        mb: 2,
                                        display: "flex",
                                        alignItems: "center",
                                        justifyContent: "space-between",
                                        gap: 1,
                                    }}
                                >
                                    <Typography variant="h6" sx={{ fontWeight: 700 }}>
                                        Жалобы
                                    </Typography>

                                    <Chip
                                        label={complaints.length}
                                        size="small"
                                        variant="outlined"
                                        color={complaints.length > 0 ? "error" : "default"}
                                    />
                                </Box>

                                <Divider sx={{ mb: 2 }} />

                                {complaints.length > 0 ? (
                                    <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
                                        {complaints.map((complaint, index) => (
                                            <Alert key={index} severity="warning">
                                                {JSON.stringify(complaint)}
                                            </Alert>
                                        ))}
                                    </Box>
                                ) : (
                                    <Typography variant="body2" sx={{ color: "text.secondary" }}>
                                        Жалоб нет
                                    </Typography>
                                )}
                            </CardContent>
                        </Card>
                    </Box>
                </Box>
            </Box>
        </Container>
    );
}

interface InfoRowProps {
    label: string;
    value: ReactNode;
}

function InfoRow({ label, value }: InfoRowProps) {
    return (
        <Box
            sx={{
                p: 1.5,
                borderRadius: 2,
                bgcolor: "background.default",
                display: "flex",
                flexDirection: { xs: "column", sm: "row" },
                alignItems: { xs: "flex-start", sm: "center" },
                justifyContent: "space-between",
                gap: 1,
            }}
        >
            <Typography variant="body2" sx={{ color: "text.secondary" }}>
                {label}
            </Typography>

            <Typography variant="body1" sx={{ fontWeight: 700 }}>
                {value}
            </Typography>
        </Box>
    );
}