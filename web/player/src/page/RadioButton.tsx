import HLS from "hls.js";
import styles from "./RadioButton.module.css";
import { setTrackStore, trackStore } from "../store/track";
import { onMount } from "solid-js";
import { addEventListener, EVENTS } from "../store/events";

const STREAM_SOURCE = "/stream";

export const RadioButton = () => {
    let videoRef: HTMLVideoElement | undefined;
    let hls: HLS | undefined;

    const initStream = () => {
        if (!trackStore.isPlay && HLS.isSupported()) {
            hls = new HLS();
            hls.loadSource(STREAM_SOURCE);
            hls.attachMedia(videoRef as unknown as HTMLMediaElement);
        }
    };

    const handlePlay = () => {
        setTrackStore("isPlay", true);
        videoRef?.play();
    };

    const handlePause = () => {
        setTrackStore("isPlay", false);
        videoRef?.pause();
        hls?.destroy();
    };

    const togglePlayback = () => {
        initStream();

        if (trackStore.isPlay) {
            handlePause();
        } else {
            handlePlay();
        }
    };

    onMount(() => {
        addEventListener(EVENTS.pause, (_e: MessageEvent<string>) => {
            setTrackStore("trackName", "");
            handlePause();
        });

        addEventListener(EVENTS.play, (e: MessageEvent<string>) => {
            setTrackStore("trackName", e.data);
            if (trackStore.isPlay) handlePause();
            initStream();
            handlePlay();
        });
    });

    return (
        <div>
            <video id="video" ref={videoRef}></video>
            <div class={styles.container}>
                <div class={`${styles.box} ${trackStore.isPlay ? styles.pause : ""}`} onClick={togglePlayback}></div>
            </div>
        </div>
    );
};
