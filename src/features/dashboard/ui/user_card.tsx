import {
    Avatar,
    Box,
    Card,
    CardContent,
    Chip,
    Divider,
    LinearProgress,
    Stack,
    Typography,
} from "@mui/material";

import type { UserType } from "../../../shared/entities/user/user_types.ts";

interface UserRatingCardProps {
    user: UserType;
    position?: number;
}

export function UserRatingCard({ user, position }: UserRatingCardProps) {
    const fullName =
        user.FullName ?? `${user.Name ?? ""} ${user.LastName ?? ""}`.trim();

    const initials = `${user.Name?.[0] ?? ""}${user.LastName?.[0] ?? ""}`.toUpperCase();

    const rating = Math.min(Math.max(Number(user.Rating ?? 0), 0), 100);
    const ratingColor = getRatingColor(rating);

    const notesCount = user.Notes?.length ?? 0;
    const complaintsCount = user.Complaints?.length ?? 0;

    return (
        <Card
            variant="outlined"
            sx={{
                height: "100%",
                borderRadius: 4,
                transition: "0.2s ease",
                "&:hover": {
                    transform: "translateY(-2px)",
                    boxShadow: 4,
                },
            }}
        >
            <CardContent>
                <Stack spacing={2.5}>
                    <Box
                        sx={{
                            display: "flex",
                            alignItems: "center",
                            gap: 2,
                        }}
                    >
                        <Avatar
                            sx={{
                                width: 56,
                                height: 56,
                                bgcolor: "primary.main",
                                fontWeight: 800,
                            }}
                        >
                            {initials || "?"}
                        </Avatar>

                        <Box sx={{ minWidth: 0, flex: 1 }}>
                            <Typography
                                variant="h6"
                                sx={{
                                    fontWeight: 800,
                                    lineHeight: 1.2,
                                }}
                                noWrap
                            >
                                {fullName || "Без имени"}
                            </Typography>

                            <Typography
                                variant="body2"
                                sx={{
                                    color: "text.secondary",
                                }}
                                noWrap
                            >
                                @{user.Login}
                            </Typography>
                        </Box>

                        {position && (
                            <Chip
                                label={`#${position}`}
                                color="primary"
                                variant="outlined"
                                size="small"
                                sx={{ fontWeight: 700 }}
                            />
                        )}
                    </Box>

                    <Stack
                        direction="row"
                        spacing={1}
                        sx={{
                            flexWrap: "wrap",
                            rowGap: 1,
                        }}
                    >
                        <Chip
                            label={getRoleLabel(user.Role)}
                            size="small"
                            color={user.Role === "User" ? "default" : "primary"}
                        />

                        <Chip
                            label={`Класс ${user.Class}`}
                            size="small"
                            variant="outlined"
                        />
                    </Stack>

                    <Box>
                        <Box
                            sx={{
                                mb: 1,
                                display: "flex",
                                alignItems: "center",
                                justifyContent: "space-between",
                                gap: 1,
                            }}
                        >
                            <Typography
                                variant="body2"
                                sx={{
                                    color: "text.secondary",
                                }}
                            >
                                Социальный рейтинг
                            </Typography>

                            <Typography
                                variant="h6"
                                sx={{
                                    fontWeight: 900,
                                }}
                            >
                                {user.Rating}
                            </Typography>
                        </Box>

                        <LinearProgress
                            variant="determinate"
                            value={rating}
                            color={ratingColor}
                            sx={{
                                height: 10,
                                borderRadius: 999,
                            }}
                        />
                    </Box>

                    <Divider />

                    <Box
                        sx={{
                            display: "grid",
                            gridTemplateColumns: "1fr 1fr",
                            gap: 1.5,
                        }}
                    >
                        <SmallStat
                            label="Заметки"
                            value={notesCount}
                        />

                        <SmallStat
                            label="Жалобы"
                            value={complaintsCount}
                            danger={complaintsCount > 0}
                        />
                    </Box>
                </Stack>
            </CardContent>
        </Card>
    );
}

interface SmallStatProps {
    label: string;
    value: number;
    danger?: boolean;
}

function SmallStat({ label, value, danger = false }: SmallStatProps) {
    return (
        <Box
            sx={{
                p: 1.5,
                borderRadius: 2,
                bgcolor: danger ? "error.light" : "background.default",
            }}
        >
            <Typography
                variant="caption"
                sx={{
                    color: danger ? "error.contrastText" : "text.secondary",
                }}
            >
                {label}
            </Typography>

            <Typography
                variant="h6"
                sx={{
                    fontWeight: 900,
                    color: danger ? "error.contrastText" : "text.primary",
                }}
            >
                {value}
            </Typography>
        </Box>
    );
}

function getRatingColor(rating: number) {
    if (rating >= 80) return "success" as const;
    if (rating >= 60) return "warning" as const;
    return "error" as const;
}

function getRoleLabel(role: string) {
    const roles: Record<string, string> = {
        Owner: "Владелец",
        Admin: "Администратор",
        Helper: "Помощник",
        User: "Ученик",
    };

    return roles[role] ?? role;
}