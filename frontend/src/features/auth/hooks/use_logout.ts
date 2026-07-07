import {useMutation, useQueryClient} from "@tanstack/react-query";
import {authApi} from "../api/auth_api.ts";
import {useAuthStore} from "../store/auth_store.ts";

export const useLogout = () => {
    const queryclient = useQueryClient();
    return useMutation({
        mutationFn: () => authApi.logout(),
        onSuccess: () => {
            useAuthStore.setState({
                user: null,
                status: "unauthenticated",
                error: null,
            });
            queryclient.invalidateQueries();
        }
    })
}