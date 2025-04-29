import { onMount, onCleanup } from "solid-js";
import { CurrentTrack } from "./CurrentTrack";
import { ListenersCounter } from "./ListenersCounter";
import { RadioButton } from "./RadioButton";
import { closeEventSource, initEventSource } from "../store/events";

export const Page = () => {
    onMount(() => {
        initEventSource();
    });

    onCleanup(() => {
        closeEventSource();
    });

    return (
        <div style={{ display: "flex", "flex-direction": "column", height: "100vh" }}>
            <ListenersCounter />
            <RadioButton />
            <CurrentTrack />
        </div>
    );
};
