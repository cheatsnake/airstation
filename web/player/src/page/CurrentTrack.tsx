import { createSignal, onMount } from "solid-js";
import { airstationAPI } from "../api";
import styles from "./CurrentTrack.module.css";
import { addEventListener, EVENTS } from "../store/events";

export const CurrentTrack = () => {
    const [track, setTrack] = createSignal("");

    onMount(async () => {
        try {
            const cs = await airstationAPI.getPlayback();
            if (cs.currentTrack) setTrack(cs.currentTrack.name);
        } catch (error) {
            console.log(error);
        }

        addEventListener(EVENTS.newTrack, (e: MessageEvent<string>) => {
            setTrack(e.data);
        });
    });

    const copyToClipboard = async () => {
        try {
            await navigator.clipboard.writeText(track());
        } catch (error) {
            console.log(error);
        }
    };

    return (
        <div class={styles.box}>
            <div onClick={copyToClipboard} class={styles.label}>
                {track()}
            </div>
        </div>
    );
};
