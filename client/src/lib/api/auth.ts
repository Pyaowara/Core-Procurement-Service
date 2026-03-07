import { request } from "./client";

export interface User {
    id: number;
    username: string;
    role: string;
    first_name: string;
    last_name: string;
    email: string;
}

export const authApi = {
    login: (data: { username: string; password: string }) =>
        request<{ message: string; token: string; user: User }>("/auth/login", {
            method: "POST",
            body: JSON.stringify(data),
        }),

    register: (data: {
        username: string;
        password: string;
        first_name: string;
        last_name: string;
        email: string;
    }) =>
        request<{ message: string; user: User }>("/auth/register", {
            method: "POST",
            body: JSON.stringify(data),
        }),

    logout: () =>
        request<{ message: string }>("/auth/logout", { method: "POST" }),

    me: () => request<User>("/auth/me"),
};
