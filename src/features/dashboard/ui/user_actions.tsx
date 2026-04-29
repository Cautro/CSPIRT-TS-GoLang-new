import { Box, Typography, Stack, TextField, Button } from "@mui/material";
import { UserRoles } from "../../../shared/entities/user/user_types.ts";
import type { UserType } from "../../../shared/entities/user/user_types.ts";
import { useState } from "react";
import { useDashboardStore } from "../store/dashboard_store.ts";
import type { RatingChangeDTO } from "../api/dashboard_api.ts";

interface UserActionsProps {
    user: UserType;
    role: string;
}

export function UserActions({ user, role }: UserActionsProps) {
    const [rating, setRating] = useState<number>(0);
    const [ratingReason, setRatingReason] = useState("");
    const [complaint, setComplaint] = useState("");
    const [note, setNote] = useState("");

    const changeRating = useDashboardStore((state) => state.changeRating);

    async function handleChangeRating() {
        if (rating === 0 || ratingReason.trim() === "") {
            setRatingReason("");
            setRating(0);
            return;
        }

        const dto: RatingChangeDTO = {
            rating: rating,
            reason: ratingReason.trim(),
            target_login: user.Login,
        };

        setRatingReason("");
        setRating(0);

        await changeRating(dto);
    }

    return (
        <Box sx={{ p: 2, border: "1px solid", borderColor: "divider", borderRadius: 2 }}>
            <Stack spacing={1}>
                <Typography variant="h6" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                    {user.Name} {user.LastName}
                </Typography>

                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    @{user.Login}
                </Typography>

                <Typography variant="body2">
                    <strong>Класс:</strong> {user.Class}
                </Typography>

                <Typography variant="body2">
                    <strong>Роль:</strong> {UserRoles[user.Role] || "Не определена"}
                </Typography>

                <Typography variant="body1" sx={{ fontWeight: 700, color: "primary.main", mt: 1 }}>
                    Рейтинг: {user.Rating}
                </Typography>

                {(role === "Admin" || role === "Owner") && (
                    <Box sx={{ display: "flex", flexDirection: "column", gap: 1, mt: 2 }}>
                        <TextField
                            label="Изменить рейтинг"
                            type="number"
                            onChange={(e) => setRating(Number(e.target.value))}
                            value={rating}
                        />

                        <TextField
                            label="Причина"
                            type="text"
                            onChange={(e) => setRatingReason(e.target.value)}
                            value={ratingReason}
                        />

                        <Button variant="contained" onClick={() => void handleChangeRating()}>
                            Изменить
                        </Button>
                    </Box>
                )}

                {role === "Helper" && (
                    <Box sx={{ display: "flex", flexDirection: "column", gap: 1, mt: 2 }}>
                        <TextField
                            label="Оставить заметку"
                            type="text"
                            onChange={(e) => setNote(e.target.value)}
                            value={note}
                        />

                        <Button variant="contained">
                            Отправить
                        </Button>
                    </Box>
                )}

                <Box sx={{ display: "flex", flexDirection: "column", gap: 1, mt: 2 }}>
                    <TextField
                        label="Жалоба"
                        type="text"
                        onChange={(e) => setComplaint(e.target.value)}
                        value={complaint}
                    />

                    <Button variant="contained">
                        Отправить
                    </Button>
                </Box>
            </Stack>
        </Box>
    );
}