import { notifications } from "@mantine/notifications";
import { handleErr } from "../utils/error";

export const errNotify = (err: string | unknown) => {
    const message = typeof err === "string" ? err : handleErr(err);

    notifications.show({
        message,
        withBorder: true,
        withCloseButton: true,
        autoClose: 20_000,
        color: "red",
    });
};

export const okNotify = (message: string) => {
    notifications.show({
        message,
        withBorder: true,
        withCloseButton: true,
        color: "green",
    });
};

export const infoNotify = (message: string) => {
    notifications.show({
        message,
        withBorder: true,
        withCloseButton: true,
        color: "blue",
    });
};

export const warnNotify = (message: string) => {
    notifications.show({
        message,
        withBorder: true,
        withCloseButton: true,
        color: "yellow",
    });
};
