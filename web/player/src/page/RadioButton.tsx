import HLS from "hls.js";
import styles from "./RadioButton.module.css";
import { setTrackStore, trackStore } from "../store/track";
import { Component, onCleanup, onMount } from "solid-js";
import { addEventListener, EVENTS } from "../store/events";
import { getUnixTime } from "../utils/date";
import { addHistory } from "../store/history";
import { getCssVariable } from "../utils/document";
import { getHueFromHex } from "../utils/color";

const STREAM_SOURCE = "/stream";

export const RadioButton = () => {
    let videoRef: HTMLAudioElement | undefined;
    let hls: HLS | undefined;
    let accentHue: number | null = null;

    const initStream = () => {
        if (!trackStore.isPlay && HLS.isSupported()) {
            hls = new HLS();
            hls.loadSource(STREAM_SOURCE);
            hls.attachMedia(videoRef as unknown as HTMLMediaElement);
        }
    };

    const handlePlay = () => {
        initStream();
        if (!trackStore.trackName) return;
        setTrackStore("isPlay", true);
    };

    const handlePause = () => {
        setTrackStore("isPlay", false);
        hls?.destroy();
    };

    onMount(() => {
        addEventListener(EVENTS.pause, (_e: MessageEvent<string>) => {
            setTrackStore("trackName", "");
            (() => videoRef?.pause())();
        });

        addEventListener(EVENTS.play, (e: MessageEvent<string>) => {
            const unixTime = getUnixTime();
            setTrackStore("trackName", e.data);
            addHistory({ id: unixTime, playedAt: unixTime, trackName: e.data });

            if (trackStore.isPlay) (() => videoRef?.pause())();
            (() => videoRef?.play())();
        });

        document.body.addEventListener("keydown", (event) => {
            if (event.key === " ") {
                event.preventDefault();
                trackStore.isPlay ? videoRef?.pause() : videoRef?.play();
            }
        });

        const accentColor = getCssVariable("--accent-color");
        if (accentColor) accentHue = getHueFromHex(accentColor);
    });

    return (
        <div class={styles.container}>
            <audio id="video" ref={videoRef} onPause={handlePause} onPlay={handlePlay}></audio>
            <div class={styles.box}>
                {trackStore.isPlay ? (
                    <AnimatedPauseButton pause={() => videoRef?.pause()} media={videoRef} accentHue={accentHue} />
                ) : (
                    <div class={styles.play_icon} tabIndex={0} role="button" onClick={() => videoRef?.play()}></div>
                )}
            </div>
        </div>
    );
};

let audioSource: MediaElementAudioSourceNode | null = null;
let audioContext: AudioContext | null = null;

const AnimatedPauseButton: Component<{ pause: () => void; accentHue: number | null; media?: HTMLAudioElement }> = (
    props,
) => {
    let pauseIconRef: HTMLDivElement | undefined;
    let analyser: AnalyserNode | null = null;
    let dataArray: Uint8Array | null = null;
    let animationId: number | null = null;
    let gainNode: GainNode | null = null;
    let currentHue = props.accentHue || 0;
    let currentSaturation = props.accentHue ? 100 : 50;
    let currentLightness = 60;

    onMount(async () => {
        if (!pauseIconRef || !props.media) return;
        await initAudio();
        draw();
    });

    onCleanup(async () => {
        if (animationId !== null) {
            cancelAnimationFrame(animationId);
            animationId = null;
        }

        if (gainNode) {
            gainNode.disconnect();
            gainNode = null;
        }

        if (analyser) {
            analyser.disconnect();
            analyser = null;
        }

        dataArray = null;

        if (pauseIconRef) {
            pauseIconRef.style.transform = "scale(1)";
            pauseIconRef.style.backgroundColor = "white";
            pauseIconRef.style.boxShadow = "none";
        }
    });

    const initAudio = async () => {
        try {
            if (!props.media) return;
            if (!audioContext) audioContext = new window.AudioContext();

            analyser = audioContext.createAnalyser();
            analyser.fftSize = 256;
            gainNode = audioContext.createGain();
            gainNode.gain.value = 1;

            if (!audioSource) audioSource = audioContext.createMediaElementSource(props.media);
            audioSource.connect(gainNode);
            gainNode.connect(analyser);
            analyser.connect(audioContext.destination);

            const bufferLength = analyser.frequencyBinCount;
            dataArray = new Uint8Array(bufferLength);
        } catch (err) {
            console.error("Error initializing audio:", err);
        }
    };

    const draw = () => {
        if (!pauseIconRef || !analyser || !dataArray) return;

        animationId = requestAnimationFrame(draw);
        analyser.getByteFrequencyData(dataArray as Uint8Array<ArrayBuffer>);

        let bass = 0;
        let treble = 0;
        const bassEnd = Math.floor(dataArray.length * 0.3);
        const trebleStart = Math.floor(dataArray.length * 0.6);

        for (let i = 0; i < dataArray.length; i++) {
            if (i < bassEnd) bass += dataArray[i];
            else if (i > trebleStart) treble += dataArray[i];
        }

        bass /= bassEnd;
        treble /= dataArray.length - trebleStart;

        const scale = 1 + bass / 300;
        const jump = (bass / 300) * 20;

        pauseIconRef.style.transform = `translateY(${-jump}px) scale(${scale})`;

        const bassImpact = bass / 255;
        const trebleImpact = treble / 255;

        if (props.accentHue == null) {
            currentHue += (Math.random() - 0.5) * bassImpact * 120;
            currentHue += trebleImpact * 2;
            currentHue = (currentHue + 360) % 360;
        }

        const color = `hsl(${currentHue}, ${currentSaturation}%, ${currentLightness}%)`;
        pauseIconRef.style.backgroundColor = color;

        const glowIntensity = bass / 2 + 20;
        pauseIconRef.style.boxShadow = `0 0 ${glowIntensity}px ${color}`;
    };

    return <div ref={pauseIconRef} tabIndex={0} role="button" class={styles.pause_icon} onClick={props.pause}></div>;
};
