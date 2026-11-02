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
const STREAM_RETRY_MS = 1200;
// Live segments arrive every two seconds, so allow a generous gap before
// assuming the media element is stuck.
const STALL_TIMEOUT_MS = 6000;
const HEALTH_INTERVAL_MS = 1000;

export const RadioButton = () => {
    let videoRef: HTMLAudioElement | undefined;
    let hls: HLS | undefined;
    let wantPlay = false;
    let retryTimer: number | undefined;
    let healthTimer: number | undefined;
    let lastMediaTime = -1;
    let lastProgressAt = 0;
    let lastRecoveryAt = 0;

    const clearRetry = () => {
        if (retryTimer !== undefined) {
            window.clearTimeout(retryTimer);
            retryTimer = undefined;
        }
    };

    const resetProgressClock = () => {
        lastMediaTime = -1;
        lastProgressAt = Date.now();
    };

    const destroyHls = () => {
        clearRetry();
        if (hls) {
            hls.destroy();
            hls = undefined;
        }
    };

    const safePlay = () => {
        if (!wantPlay || !videoRef || !videoRef.paused) return;
        const promise = videoRef.play();
        // A manifest swap can make play() reject before the new MediaSource is
        // ready. The readiness handlers and the health loop retry it.
        if (promise) void promise.catch(() => undefined);
    };

    const scheduleReload = () => {
        if (!wantPlay || retryTimer !== undefined) return;
        retryTimer = window.setTimeout(() => {
            retryTimer = undefined;
            if (wantPlay) attachStream(true);
        }, STREAM_RETRY_MS);
    };

    const handleHlsError = (instance: HLS, data: { fatal: boolean; type: string }) => {
        if (!data.fatal) return;

        if (data.type === HLS.ErrorTypes.MEDIA_ERROR) {
            instance.recoverMediaError();
        }

        // Live and automation are separate playlists with independent media
        // sequence numbers. When the station switches modes the old buffer is
        // useless, so rebuild the HLS instance from a fresh manifest.
        scheduleReload();
    };

    const attachStream = (reload: boolean) => {
        if (!videoRef) return;
        wantPlay = true;

        if (reload) {
            // Drop the old MediaSource before loading the new manifest. Live and
            // automation playlists have unrelated media sequences, so keeping
            // the old buffer makes hls.js wait for fragments that will never
            // arrive.
            videoRef.pause();
            destroyHls();
            videoRef.removeAttribute("src");
            videoRef.load();
            resetProgressClock();
        }

        if (HLS.isSupported()) {
            if (!hls) {
                const instance = new HLS({
                    enableWorker: true,
                    manifestLoadingMaxRetry: 8,
                    levelLoadingMaxRetry: 8,
                    fragLoadingMaxRetry: 8,
                });
                hls = instance;

                // Ignore late events from an instance that was already replaced
                // by a mode switch. Otherwise a delayed fatal error from the old
                // live playlist can tear down the healthy automation stream.
                const isCurrent = () => hls === instance;

                instance.on(HLS.Events.MEDIA_ATTACHED, () => {
                    if (!isCurrent()) return;
                    // Bust any HTTP cache for the manifest. Segment names are
                    // unique per track / live session, so only the playlist
                    // needs this.
                    instance.loadSource(`${STREAM_SOURCE}?v=${Date.now()}`);
                    safePlay();
                });
                instance.on(HLS.Events.MANIFEST_PARSED, () => {
                    if (isCurrent()) safePlay();
                });
                instance.on(HLS.Events.LEVEL_LOADED, () => {
                    if (isCurrent()) safePlay();
                });
                instance.on(HLS.Events.FRAG_BUFFERED, () => {
                    if (isCurrent()) safePlay();
                });
                instance.on(HLS.Events.ERROR, (_event, data) => {
                    if (isCurrent()) handleHlsError(instance, data);
                });
                instance.attachMedia(videoRef);
            }
        } else if (videoRef.canPlayType("application/vnd.apple.mpegurl")) {
            if (reload || !videoRef.src) {
                videoRef.src = `${STREAM_SOURCE}?v=${Date.now()}`;
                videoRef.load();
                resetProgressClock();
            }
        }

        safePlay();
    };

    const handlePlay = () => {
        setTrackStore("isPlay", true);
        attachStream(false);
    };

    const handlePause = () => {
        wantPlay = false;
        videoRef?.pause();
        destroyHls();
        setTrackStore("isPlay", false);
    };

    const handleSpace = (event: KeyboardEvent) => {
        if (event.key !== " ") return;
        event.preventDefault();
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
            const unixTime = getUnixTime();
            setTrackStore("trackName", e.data);
            addHistory({ id: unixTime, playedAt: unixTime, trackName: e.data });
            // This event also fires when the station switches between live and
            // automation, so always attach a fresh HLS session.
            attachStream(true);
        });

        // Last line of defence: if the media element stops advancing while it
        // is supposed to be playing (a missed SSE event, a rejected play()
        // promise, or a stale manifest whose sequence never catches up), rebuild
        // the stream automatically instead of requiring a pause/play click.
        healthTimer = window.setInterval(() => {
            if (!wantPlay || !videoRef) return;

            if (videoRef.paused) {
                lastProgressAt = Date.now();
                safePlay();
                return;
            }

            const now = Date.now();
            if (videoRef.currentTime !== lastMediaTime) {
                lastMediaTime = videoRef.currentTime;
                lastProgressAt = now;
                return;
            }

            if (now - lastProgressAt < STALL_TIMEOUT_MS) return;
            if (now - lastRecoveryAt < STALL_TIMEOUT_MS) return;

            lastRecoveryAt = now;
            lastProgressAt = now;
            attachStream(true);
        }, HEALTH_INTERVAL_MS);

        document.body.addEventListener("keydown", handleSpace);
    });

    onCleanup(() => {
        document.body.removeEventListener("keydown", handleSpace);
        if (healthTimer !== undefined) window.clearInterval(healthTimer);
        wantPlay = false;
        destroyHls();
    });

    return (
        <div class={styles.container}>
            <audio
                id="video"
                ref={videoRef}
                onPause={() => {
                    if (!wantPlay) setTrackStore("isPlay", false);
                }}
                onPlay={() => {
                    setTrackStore("isPlay", true);
                    resetProgressClock();
                }}
                onCanPlay={safePlay}
                onLoadedData={safePlay}
                onPlaying={safePlay}
                onError={() => {
                    if (wantPlay) scheduleReload();
                }}
            ></audio>
            <div class={styles.box}>
                {trackStore.isPlay ? (
                    <AnimatedPauseButton pause={handlePause} media={videoRef} />
                ) : (
                    <div class={styles.play_icon} tabIndex={0} role="button" onClick={handlePlay}></div>
                )}
            </div>
        </div>
    );
};

let audioSource: MediaElementAudioSourceNode | null = null;
let audioContext: AudioContext | null = null;

const AnimatedPauseButton: Component<{ pause: () => void; media?: HTMLAudioElement }> = (props) => {
    let pauseIconRef: HTMLDivElement | undefined;
    let analyser: AnalyserNode | null = null;
    let dataArray: Uint8Array | null = null;
    let animationId: number | null = null;
    let gainNode: GainNode | null = null;
    let accentHue: number | null = null;
    let currentHue = 0;
    let currentSaturation = 50;
    let currentLightness = 60;

    const loadAccentColor = () => {
        const accentColor = getCssVariable("--accent-color");
        accentHue = accentColor ? getHueFromHex(accentColor) : null;

        currentHue = accentHue !== null ? accentHue : 0;
        currentSaturation = accentHue !== null ? 100 : 50;
    };

    onMount(async () => {
        loadAccentColor();
        setInterval(loadAccentColor, 1000); // Need for hot reload

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

        if (accentHue == null) {
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
