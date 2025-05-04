import { onMount, onCleanup } from "solid-js";
import { CurrentTrack } from "./CurrentTrack";
import { ListenersCounter } from "./ListenersCounter";
import { RadioButton } from "./RadioButton";
import { closeEventSource, initEventSource } from "../store/events";
import { History } from "./History";
import styles from "./Page.module.css";

export const Page = () => {
    onMount(() => {
        initEventSource();
    });

    onCleanup(() => {
        closeEventSource();
    });

    return (
        <div class={styles.page}>
            <div class={styles.header}>
                <History />
                <ListenersCounter />
                <div class="empty_icon"> </div>
            </div>
            <RadioButton />
            <CurrentTrack />
        </div>
    );
};
