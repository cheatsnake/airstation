import { createSignal } from "solid-js";
import HLS from "hls.js";
import styles from "./RadioButton.module.css";

const STREAM_SOURCE = "http://localhost:7331/stream";

export const RadioButton = () => {
    let videoRef: HTMLVideoElement | undefined;
    let hls: HLS | undefined;

    const [isPlay, setIsPlay] = createSignal(false);

    const togglePlayback = () => {
        if (!isPlay() && HLS.isSupported()) {
            hls = new HLS();
            hls.loadSource(STREAM_SOURCE);
            hls.attachMedia(videoRef as unknown as HTMLMediaElement);
        }

        if (isPlay()) {
            setIsPlay(false);
            videoRef?.pause();
            hls?.destroy();
        } else {
            setIsPlay(true);
            videoRef?.play();
        }
    };

    return (
        <div>
            <video id="video" ref={videoRef}></video>
            <div class={styles.container}>
                <div class={`${styles.box} ${isPlay() ? styles.pause : ""}`} onClick={togglePlayback}></div>
            </div>
        </div>
    );
};
