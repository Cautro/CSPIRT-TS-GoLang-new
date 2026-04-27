import { type FormEvent, useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    CircularProgress,
    Container,
    Paper,
    TextField,
    Typography,
} from "@mui/material";

import { useAuthStore } from "../store/auth_store";

export function LoginPage() {
    const navigate = useNavigate();

    const login = useAuthStore((state) => state.login);
    const token = useAuthStore((state) => state.token);
    const status = useAuthStore((state) => state.status);
    const error = useAuthStore((state) => state.error);

    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");

    const isLoading = status === "loading";
    const isSubmitDisabled =
        isLoading || username.trim().length === 0 || password.trim().length === 0;

    if (token && status === "authenticated") {
        return <Navigate to="/" replace />;
    }

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        const success = await login({
            login: username.trim(),
            password,
        });

        if (success) {
            navigate("/", { replace: true });
        }
    }

    return (
        <Container maxWidth="sm">
            <Box
                sx={{
                    minHeight: "100vh",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    py: 4,
                }}
            >
                <Paper
                    elevation={0}
                    sx={{
                        width: "100%",
                        borderRadius: 4,
                        border: "1px solid",
                        borderColor: "divider",
                        overflow: "hidden",
                    }}
                >
                    <Card
                        variant="outlined"
                        sx={{
                            border: "none",
                            borderRadius: 0,
                        }}
                    >
                        <CardContent
                            sx={{
                                p: {
                                    xs: 3,
                                    sm: 4,
                                },
                            }}
                        >
                            <Box
                                sx={{
                                    mb: 3,
                                }}
                            >
                                <Typography
                                    variant="h4"
                                    sx={{
                                        fontWeight: 800,
                                        mb: 1,
                                    }}
                                >
                                    Вход
                                </Typography>

                                <Typography
                                    variant="body2"
                                    sx={{
                                        color: "text.secondary",
                                    }}
                                >
                                    Введите логин и пароль для доступа к профилю.
                                </Typography>
                            </Box>

                            {error && (
                                <Alert
                                    severity="error"
                                    sx={{
                                        mb: 2,
                                    }}
                                >
                                    {error}
                                </Alert>
                            )}

                            <Box
                                component="form"
                                onSubmit={handleSubmit}
                                sx={{
                                    display: "flex",
                                    flexDirection: "column",
                                    gap: 2,
                                }}
                            >
                                <TextField
                                    label="Логин"
                                    value={username}
                                    onChange={(event) => setUsername(event.target.value)}
                                    placeholder="Например: Owner"
                                    type="text"
                                    autoComplete="username"
                                    fullWidth
                                    disabled={isLoading}
                                />

                                <TextField
                                    label="Пароль"
                                    value={password}
                                    onChange={(event) => setPassword(event.target.value)}
                                    placeholder="Введите пароль"
                                    type="password"
                                    autoComplete="current-password"
                                    fullWidth
                                    disabled={isLoading}
                                />

                                <Button
                                    type="submit"
                                    variant="contained"
                                    size="large"
                                    disabled={isSubmitDisabled}
                                    sx={{
                                        mt: 1,
                                        py: 1.2,
                                        borderRadius: 2,
                                        fontWeight: 700,
                                        textTransform: "none",
                                    }}
                                >
                                    {isLoading ? (
                                        <Box
                                            sx={{
                                                display: "flex",
                                                alignItems: "center",
                                                justifyContent: "center",
                                                gap: 1,
                                            }}
                                        >
                                            <CircularProgress size={20} color="inherit" />
                                            <span>Вход...</span>
                                        </Box>
                                    ) : (
                                        "Войти"
                                    )}
                                </Button>
                            </Box>

                            <Box
                                sx={{
                                    mt: 3,
                                    p: 2,
                                    borderRadius: 2,
                                    bgcolor: "background.default",
                                }}
                            >
                                <Typography
                                    variant="body2"
                                    sx={{
                                        color: "text.secondary",
                                    }}
                                >
                                    Тестовый аккаунт:
                                </Typography>

                                <Typography
                                    variant="body2"
                                    sx={{
                                        mt: 0.5,
                                        fontWeight: 700,
                                    }}
                                >
                                    Login: Owner
                                </Typography>
                            </Box>
                        </CardContent>
                    </Card>
                </Paper>
            </Box>
        </Container>
    );
}