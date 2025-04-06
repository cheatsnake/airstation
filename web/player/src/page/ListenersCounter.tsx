import { createSignal, onMount } from "solid-js";
import { addEventListener, EVENTS } from "../store/events";
import styles from "./ListenersCounter.module.css";

export const ListenersCounter = () => {
    const [count, setCount] = createSignal(1);

    onMount(() => {
        addEventListener(EVENTS.countListeners, (e: MessageEvent<string>) => {
            setCount(+e.data);
        });
    });

    return (
        <div class={styles.counter}>
            <div class={styles.box}>
                <div class={styles.icon}></div>
                <div class={styles.number}>{count()}</div>
            </div>
        </div>
    );
};
