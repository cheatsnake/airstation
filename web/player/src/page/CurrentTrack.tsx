import { onCleanup, onMount, Show } from "solid-js";
import { airstationAPI } from "../api";
import { PlaybackState } from "../api/types";
import styles from "./CurrentTrack.module.css";
import { addEventListener, EVENTS } from "../store/events";
import { setTrackStore, trackStore } from "../store/track";
import { addHistory } from "../store/history";
import { getUnixTime } from "../utils/date";

const METADATA_REFRESH_MS = 5000;

export const CurrentTrack = () => {
    const applyPlayback = (pb: PlaybackState) => {
        setTrackStore({
            trackName: pb.currentTrack?.name ?? "",
            nextTrackName: pb.nextTrack?.name ?? "",
            voiceOverName: pb.currentVoiceOver?.voiceName ?? "",
        });
    };

    const refresh = async () => {
        try {
            applyPlayback(await airstationAPI.getPlayback());
        } catch (error) {
            console.log(error);
        }
    };

    onMount(() => {
        let interval: number | undefined;

        void refresh();

        addEventListener(EVENTS.newTrack, (e: MessageEvent<string>) => {
            const unixTime = getUnixTime();
            setTrackStore("trackName", e.data);
            addHistory({ id: unixTime, playedAt: unixTime, trackName: e.data });
            void refresh();
        });

        // A play event is also emitted when the station returns from live to
        // automation. Refresh immediately so the name changes even before the
        // first new_track event.
        addEventListener(EVENTS.play, (e: MessageEvent<string>) => {
            if (e.data) setTrackStore("trackName", e.data);
            void refresh();
        });

        addEventListener(EVENTS.pause, () => {
            setTrackStore({ trackName: "", nextTrackName: "", voiceOverName: "" });
        });

        // Polling keeps the metadata correct even if an SSE event is missed
        // during a reconnect or a long live segment.
        interval = window.setInterval(() => {
            if (!document.hidden) void refresh();
        }, METADATA_REFRESH_MS);

        const handleVisibility = () => {
            if (!document.hidden) void refresh();
        };
        document.addEventListener("visibilitychange", handleVisibility);

        onCleanup(() => {
            if (interval !== undefined) window.clearInterval(interval);
            document.removeEventListener("visibilitychange", handleVisibility);
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
                <div class={styles.now_line}>
                    <div onClick={copyToClipboard} class={styles.label}>
                        {trackStore.trackName}
                    </div>
                    <Show when={trackStore.voiceOverName.length > 0}>
                        <div class={styles.voice_badge}>🎙 {trackStore.voiceOverName}</div>
                    </Show>
                </div>
                <Show when={trackStore.nextTrackName.length > 0}>
                    <div class={styles.next_line}>Up next · {trackStore.nextTrackName}</div>
                </Show>
            </Show>
        </div>
    );
};

const OfflineLabel = () => {
    return (
        <div class={styles.offline_label}>
            <div class={styles.offline_label_icon}></div>
            <div class={styles.offline_label_title}>Stream offline</div>
        </div>
    );
};
