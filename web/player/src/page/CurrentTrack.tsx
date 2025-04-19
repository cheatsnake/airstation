import { onMount, Show } from "solid-js";
import { airstationAPI } from "../api";
import styles from "./CurrentTrack.module.css";
import { addEventListener, EVENTS } from "../store/events";
import { setTrackStore, trackStore } from "../store/track";

export const CurrentTrack = () => {
    onMount(async () => {
        try {
            const cs = await airstationAPI.getPlayback();
            if (cs.isPlaying && cs.currentTrack) setTrackStore("trackName", cs.currentTrack.name);
        } catch (error) {
            console.log(error);
        }

        addEventListener(EVENTS.newTrack, (e: MessageEvent<string>) => {
            setTrackStore("trackName", e.data);
        });
    });

    const copyToClipboard = async () => {
        try {
            await navigator.clipboard.writeText(trackStore.trackName);
        } catch (error) {
            console.log(error);
        }
    };

    return (
        <div class={styles.box}>
            <Show when={trackStore.trackName.length > 0} fallback={<OfflineLabel />}>
                <div onClick={copyToClipboard} class={styles.label}>
                    {trackStore.trackName}
                </div>
            </Show>
        </div>
    );
};

const OfflineLabel = () => {
    return <div class={styles.offline_label}>Stream offline</div>;
};
