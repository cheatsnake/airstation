import { FC, useEffect, useRef, useState } from "react";
import {
    ActionIcon,
    Badge,
    Box,
    Button,
    Card,
    Flex,
    Group,
    NumberInput,
    Paper,
    Progress,
    SegmentedControl,
    Select,
    SimpleGrid,
    Stack,
    Switch,
    Text,
    TextInput,
    Title,
    Tooltip,
} from "@mantine/core";
import { airstationAPI } from "../api";
import { modals } from "@mantine/modals";
import { formatTime } from "../utils/time";
import { usePlaybackStore } from "../store/playback";
import { useTrackQueueStore } from "../store/track-queue";
import { useCartsStore } from "../store/carts";
import { useTracksStore } from "../store/tracks";
import { useVoiceOverStore } from "../store/voiceovers";
import { useScheduleStore } from "../store/schedule";
import { KindBadge } from "../components/KindBadge";
import { ListenerCounter } from "../components/ListenerCounter";
import { EVENTS, useEventSourceStore } from "../store/events";
import { errNotify, okNotify, warnNotify } from "../notifications";
import { IconPlayerPlayFilled, IconQueue, IconReload, IconTrash } from "../icons";
import { VoiceOver } from "../api/types";
import styles from "./styles.module.css";

const DUCK_OPTIONS = [
    { value: "-6", label: "-6 dB (light)" },
    { value: "-9", label: "-9 dB" },
    { value: "-12", label: "-12 dB (default)" },
    { value: "-15", label: "-15 dB" },
    { value: "-18", label: "-18 dB" },
    { value: "-24", label: "-24 dB (deep)" },
];

export const Console: FC<{ isMobile?: boolean }> = ({ isMobile }) => {
    const playback = usePlaybackStore((s) => s.playback);
    const fetchPlayback = usePlaybackStore((s) => s.fetchPlayback);
    const queue = useTrackQueueStore((s) => s.queue);
    const fetchQueue = useTrackQueueStore((s) => s.fetchQueue);
    const fetchCarts = useCartsStore((s) => s.fetchCarts);
    const voiceOvers = useVoiceOverStore((s) => s.voiceOvers);
    const fetchVoiceOvers = useVoiceOverStore((s) => s.fetchVoiceOvers);
    const rules = useScheduleStore((s) => s.rules);
    const fetchRules = useScheduleStore((s) => s.fetchRules);
    const addEventHandler = useEventSourceStore((s) => s.addEventHandler);

    const [liveMode, setLiveMode] = useState(false);
    const [liveSegments, setLiveSegments] = useState(0);
    const [liveBusy, setLiveBusy] = useState(false);
    const [liveConnecting, setLiveConnecting] = useState(false);
    const [talking, setTalking] = useState(false);
    const [openMic, setOpenMic] = useState(false);
    const [externalUrl, setExternalUrl] = useState("");
    const [audioBitRate, setAudioBitRate] = useState<string>("48");
    const [bitRateBusy, setBitRateBusy] = useState(false);
    const [liveAnalyser, setLiveAnalyser] = useState<AnalyserNode | null>(null);

    const liveRecorderRef = useRef<MediaRecorder | null>(null);
    const liveAudioCtxRef = useRef<AudioContext | null>(null);
    const liveGainRef = useRef<GainNode | null>(null);
    const liveAnalyserRef = useRef<AnalyserNode | null>(null);
    const liveStreamRef = useRef<MediaStream | null>(null);
    const liveSocketRef = useRef<WebSocket | null>(null);
    const liveStoppingRef = useRef(false);
    const openMicRef = useRef(false);
    const talkingRef = useRef(false);

    useEffect(() => {
        void fetchPlayback();
        void fetchQueue();
        void fetchCarts();
        void fetchVoiceOvers();
        void fetchRules();

        addEventHandler(EVENTS.newTrack, () => {
            void fetchPlayback();
            void fetchQueue();
        });
        addEventHandler(EVENTS.interlude, () => {
            void fetchQueue();
            void fetchPlayback();
        });
        addEventHandler(EVENTS.voiceOver, () => {
            void fetchPlayback();
            void fetchVoiceOvers();
        });
        addEventHandler(EVENTS.loadedTracks, () => {
            void fetchCarts();
        });
    }, []);

    useEffect(() => {
        airstationAPI
            .getStationInfo()
            .then((info) => {
                if (info.audioBitRate) setAudioBitRate(String(info.audioBitRate));
            })
            .catch(() => {});
    }, []);

    useEffect(() => {
        let interval: number | undefined;
        const poll = async () => {
            try {
                const status = await airstationAPI.getLiveStatus();
                setLiveMode(status.live);
                setLiveSegments(status.segmentCount);
            } catch {
                // ignore
            }
        };
        void poll();
        interval = window.setInterval(poll, 2000);
        return () => {
            if (interval) window.clearInterval(interval);
        };
    }, []);

    const current = playback.currentTrack;
    const nextInQueue = queue.filter((t) => t.id !== current?.id).slice(0, 5);
    const progress = current ? Math.min(100, (playback.currentTrackElapsed / (current.duration || 1)) * 100) : 0;

    const saveBitRate = async () => {
        setBitRateBusy(true);
        try {
            const current = await airstationAPI.getStationInfo();
            await airstationAPI.editStationInfo({ ...current, audioBitRate: Number(audioBitRate) });
            okNotify(`Audio bitrate set to ${audioBitRate} kbps (applies to new conversions)`);
        } catch (error) {
            errNotify(error);
        } finally {
            setBitRateBusy(false);
        }
    };

    const cleanupLiveSession = () => {
        const socket = liveSocketRef.current;
        liveSocketRef.current = null;
        if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
            socket.close(1000, "Live stopped");
        }
        const recorder = liveRecorderRef.current;
        liveRecorderRef.current = null;
        if (recorder && recorder.state !== "inactive") {
            try {
                recorder.stop();
            } catch {
                // Already stopped.
            }
        }
        liveStreamRef.current?.getTracks().forEach((t) => t.stop());
        liveStreamRef.current = null;
        const ctx = liveAudioCtxRef.current;
        liveAudioCtxRef.current = null;
        liveGainRef.current = null;
        liveAnalyserRef.current?.disconnect();
        liveAnalyserRef.current = null;
        setLiveAnalyser(null);
        if (ctx && ctx.state !== "closed") {
            void ctx.close();
        }
        openMicRef.current = false;
        talkingRef.current = false;
        setOpenMic(false);
        setTalking(false);
        setLiveMode(false);
        setLiveSegments(0);
    };

    const applyMicGain = (open: boolean) => {
        const ctx = liveAudioCtxRef.current;
        const gain = liveGainRef.current;
        if (!ctx || !gain) return;
        gain.gain.setTargetAtTime(open ? 1 : 0, ctx.currentTime, 0.008);
    };

    const talk = (on: boolean) => {
        talkingRef.current = on;
        applyMicGain(on || openMicRef.current);
    };

    const toggleOpenMic = () => {
        const next = !openMicRef.current;
        openMicRef.current = next;
        setOpenMic(next);
        applyMicGain(next || talkingRef.current);
    };

    const startLiveMic = async () => {
        if (liveBusy || liveConnecting || liveMode) return;
        if (!window.isSecureContext) {
            errNotify(new Error("Microphone access requires HTTPS or localhost"));
            return;
        }

        setLiveBusy(true);
        setLiveConnecting(true);
        try {
            // Verify the JWT session before asking for the microphone. If this
            // fails, the WebSocket handshake would fail too and the browser
            // hides the HTTP status from us.
            const status = await airstationAPI.getLiveStatus();
            if (status.live) {
                errNotify(new Error(`Station is already live (${status.mode || "unknown"}). Stop it first.`));
                setLiveConnecting(false);
                setLiveBusy(false);
                return;
            }

            const stream = await navigator.mediaDevices.getUserMedia({
                audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true },
            });
            liveStreamRef.current = stream;

            const ctx = new AudioContext({ latencyHint: "interactive" });
            if (ctx.state === "suspended") {
                await ctx.resume();
            }
            liveAudioCtxRef.current = ctx;

            // The microphone runs continuously; push-to-talk only moves the
            // gain, so the WebM pipe never stalls and listeners never rebuffer.
            const source = ctx.createMediaStreamSource(stream);
            const gain = ctx.createGain();
            gain.gain.value = 0;
            liveGainRef.current = gain;
            const dest = ctx.createMediaStreamDestination();

            // Tap the post-gain signal for the on-screen waveform. The analyser
            // branch ends in a silent gain so it never plays the mic back
            // through the studio speakers.
            const analyser = ctx.createAnalyser();
            analyser.fftSize = 1024;
            analyser.smoothingTimeConstant = 0.75;
            const silentMonitor = ctx.createGain();
            silentMonitor.gain.value = 0;

            source.connect(gain);
            gain.connect(dest);
            gain.connect(analyser);
            analyser.connect(silentMonitor);
            silentMonitor.connect(ctx.destination);

            liveAnalyserRef.current = analyser;
            setLiveAnalyser(analyser);

            const mime = MediaRecorder.isTypeSupported("audio/webm;codecs=opus")
                ? "audio/webm;codecs=opus"
                : "audio/webm";
            const recorder = new MediaRecorder(dest.stream, { mimeType: mime });
            liveRecorderRef.current = recorder;

            // Every recorder chunk is appended to the same logical WebM stream on
            // the server, so FFmpeg sees one continuous input.
            recorder.ondataavailable = (event) => {
                if (!event.data || event.data.size === 0) return;
                const socket = liveSocketRef.current;
                if (!socket || socket.readyState !== WebSocket.OPEN) return;
                if (socket.bufferedAmount > 4 * 1024 * 1024) return;
                socket.send(event.data);
            };

            const socket = airstationAPI.liveIngestSocket();
            const socketUrl = airstationAPI.liveIngestUrl();
            liveSocketRef.current = socket;
            liveStoppingRef.current = false;

            // A reverse proxy that does not forward Upgrade headers leaves the
            // socket stuck in CONNECTING until the browser's own timeout.
            let connectTimedOut = false;
            const connectTimer = window.setTimeout(() => {
                if (socket.readyState === WebSocket.CONNECTING) {
                    connectTimedOut = true;
                    socket.close();
                }
            }, 10000);

            socket.onopen = () => {
                window.clearTimeout(connectTimer);
                recorder.start(250);
                setLiveConnecting(false);
                setLiveBusy(false);
                setLiveMode(true);
                setLiveSegments(0);
                setOpenMic(false);
                openMicRef.current = false;
                setTalking(false);
                talkingRef.current = false;
                okNotify("Live microphone connected — hold Talk or enable Open mic");
            };
            socket.onclose = (event) => {
                window.clearTimeout(connectTimer);
                const wasStopping = liveStoppingRef.current;
                liveStoppingRef.current = false;
                cleanupLiveSession();
                setLiveConnecting(false);
                setLiveBusy(false);
                if (wasStopping || event.code === 1000) return;
                if (connectTimedOut) {
                    errNotify(
                        new Error(
                            `Live WebSocket did not open at ${socketUrl}. If you use a reverse proxy, forward Upgrade/Connection headers.`,
                        ),
                    );
                } else if (event.code === 1006) {
                    errNotify(
                        new Error(
                            `Live WebSocket was rejected at ${socketUrl} (code 1006). Rebuild the container image so the server accepts GET upgrades.`,
                        ),
                    );
                } else {
                    errNotify(new Error(event.reason || `Live microphone disconnected (code ${event.code})`));
                }
            };
            socket.onerror = () => {
                // onclose follows with the useful reason and performs cleanup.
            };
        } catch (error) {
            errNotify(error);
            cleanupLiveSession();
            setLiveConnecting(false);
        } finally {
            setLiveBusy(false);
        }
    };

    const startLiveExternal = async () => {
        if (!externalUrl.trim()) {
            warnNotify("Enter a stream URL first");
            return;
        }
        setLiveBusy(true);
        try {
            await airstationAPI.startLiveExternal(externalUrl.trim());
            setLiveMode(true);
            setLiveSegments(0);
            okNotify("External live source started");
        } catch (error) {
            errNotify(error);
        } finally {
            setLiveBusy(false);
        }
    };

    const stopLive = async () => {
        setLiveBusy(true);
        liveStoppingRef.current = true;
        try {
            cleanupLiveSession();
            await airstationAPI.stopLive();
            okNotify("Live stopped — automation resumed");
        } catch (error) {
            errNotify(error);
        } finally {
            setLiveBusy(false);
        }
    };

    useEffect(() => {
        if (!liveMode) return;

        const isEditable = (target: EventTarget | null) => {
            const el = target as HTMLElement | null;
            if (!el) return false;
            return el.tagName === "INPUT" || el.tagName === "TEXTAREA" || el.tagName === "SELECT" || el.isContentEditable;
        };
        const down = (event: KeyboardEvent) => {
            if (event.code !== "Space" || event.repeat || isEditable(event.target)) return;
            event.preventDefault();
            setTalking(true);
            talk(true);
        };
        const up = (event: KeyboardEvent) => {
            if (event.code !== "Space" || isEditable(event.target)) return;
            event.preventDefault();
            setTalking(false);
            talk(false);
        };
        window.addEventListener("keydown", down);
        window.addEventListener("keyup", up);
        return () => {
            window.removeEventListener("keydown", down);
            window.removeEventListener("keyup", up);
        };
        // talk() only touches refs, so it does not need to be a dependency.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [liveMode]);

    const recordVoiceClip = async () => {
        setLiveBusy(true);
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            const recorder = new MediaRecorder(stream, { mimeType: "audio/webm;codecs=opus" });
            const blobs: Blob[] = [];
            recorder.ondataavailable = (e) => {
                if (e.data && e.data.size > 0) blobs.push(e.data);
            };
            const finished = new Promise<Blob>((resolve) => {
                recorder.onstop = () => resolve(new Blob(blobs, { type: "audio/webm" }));
            });
            recorder.start();
            // Record for 8 seconds then stop.
            await new Promise((resolve) => setTimeout(resolve, 8000));
            recorder.stop();
            stream.getTracks().forEach((t) => t.stop());
            const blob = await finished;
            const file = new File([blob], "voice-link.webm", { type: "audio/webm" });
            const { message } = await useTracksStore.getState().uploadTracks([file], "voice");
            okNotify(message);
            await fetchCarts();
        } catch (error) {
            errNotify(error);
        } finally {
            setLiveBusy(false);
        }
    };

    return (
        <Box h={isMobile ? "calc(100vh - 70px)" : "100%"} style={{ overflowY: "auto" }} pr={4}>
            <Stack gap="md">
                <Paper p="md" radius="md" className={styles.transparent_paper}>
                    <Flex justify="space-between" align="flex-start" gap="md">
                        <Stack gap={4} style={{ minWidth: 0 }}>
                            <Group gap="xs">
                                {playback.isPlaying ? <span className={styles.onair_dot} /> : null}
                                <Text fw={800} size="xl" c={playback.isPlaying ? "red" : "dimmed"}>
                                    {playback.isPlaying ? "ON AIR" : "OFF AIR"}
                                </Text>
                                <Badge variant="light" color={playback.isPlaying ? "red" : "gray"}>
                                    {playback.mode ?? "automation"}
                                </Badge>
                            </Group>
                            <Group gap="xs" wrap="nowrap">
                                <Title order={2} lineClamp={1}>
                                    {current?.name ?? "Nothing playing"}
                                </Title>
                                <KindBadge kind={current?.kind} size="md" />
                            </Group>
                            {playback.currentVoiceOver ? (
                                <Badge color="green" variant="light" leftSection="🎙">
                                    {`Live link over this track: ${playback.currentVoiceOver.voiceName ?? "voice"} • duck ${
                                        playback.currentVoiceOver.duckDb
                                    } dB`}
                                </Badge>
                            ) : (
                                <Text size="sm" c="dimmed">
                                    {current ? "Playing the clean music bed" : "Start playback to put the station on air"}
                                </Text>
                            )}
                        </Stack>

                        <Stack gap={0} align="flex-end">
                            <Text size="xs" c="dimmed" tt="uppercase">
                                Listeners
                            </Text>
                            <ListenerCounter large />
                        </Stack>
                    </Flex>

                    <Progress value={progress} size="lg" radius="xl" mt="sm" />
                    <Group justify="space-between" mt={6}>
                        <Text size="sm" c="dimmed">
                            {`${formatTime(playback.currentTrackElapsed)} / ${formatTime(current?.duration ?? 0)}`}
                        </Text>
                        <Text size="sm" c="dimmed" lineClamp={1}>
                            {playback.nextTrack ? `Up next: ${playback.nextTrack.name}` : "Up next: queue is empty"}
                        </Text>
                    </Group>
                </Paper>

                <Paper p="md" radius="md" className={styles.transparent_paper}>
                    <Group justify="space-between" mb="xs">
                        <Title order={4}>Live mode</Title>
                        <Badge variant="light" color={liveMode ? "red" : liveConnecting ? "orange" : "gray"}>
                            {liveMode ? `LIVE • ${liveSegments} segment(s)` : liveConnecting ? "connecting…" : "automation"}
                        </Badge>
                    </Group>
                    {liveMode ? (
                        <Stack gap="xs">
                            <Text size="sm" c="dimmed">
                                You are on air. Listeners hear you a few seconds later. Hold Talk (or Space) to
                                speak, or enable Open mic for a continuous show. Automation is paused until you
                                stop.
                            </Text>
                            <Button
                                color={talking || openMic ? "red" : "gray"}
                                size="xl"
                                h={92}
                                fullWidth
                                onPointerDown={(event) => {
                                    event.preventDefault();
                                    event.currentTarget.setPointerCapture(event.pointerId);
                                    setTalking(true);
                                    talk(true);
                                }}
                                onPointerUp={() => {
                                    setTalking(false);
                                    talk(false);
                                }}
                                onPointerCancel={() => {
                                    setTalking(false);
                                    talk(false);
                                }}
                            >
                                {openMic ? "MIC OPEN" : talking ? "ON AIR — release to mute" : "HOLD TO TALK"}
                            </Button>
                            <Box>
                                <Group justify="space-between" mb={4}>
                                    <Text size="sm" fw={600}>
                                        Live voice
                                    </Text>
                                    <Badge size="sm" variant="light" color={talking || openMic ? "red" : "gray"}>
                                        {openMic ? "mic open" : talking ? "on air" : "muted"}
                                    </Badge>
                                </Group>
                                <LiveWaveform analyser={liveAnalyser} live={liveMode} muted={!talking && !openMic} />
                            </Box>
                            <Group justify="space-between" gap="xs">
                                <Switch checked={openMic} onChange={toggleOpenMic} label="Open mic (no push-to-talk)" />
                                <Button variant="light" loading={liveBusy} onClick={recordVoiceClip}>
                                    Record 8s voice clip
                                </Button>
                            </Group>
                            <Button color="red" loading={liveBusy} onClick={stopLive}>
                                Stop live & resume automation
                            </Button>
                        </Stack>
                    ) : (
                        <Stack gap="xs">
                            <Group gap="xs">
                                <Button
                                    color="red"
                                    variant="light"
                                    loading={liveBusy || liveConnecting}
                                    onClick={startLiveMic}
                                >
                                    Go live (microphone)
                                </Button>
                                <Button variant="light" loading={liveBusy} onClick={recordVoiceClip}>
                                    Record 8s voice clip
                                </Button>
                            </Group>
                            <Group gap="xs" align="flex-end">
                                <TextInput
                                    style={{ flex: 1 }}
                                    placeholder="External stream URL (http/rtmp/file)"
                                    value={externalUrl}
                                    onChange={(e) => setExternalUrl(e.currentTarget.value)}
                                />
                                <Button variant="light" loading={liveBusy} onClick={startLiveExternal}>
                                    Go live (external)
                                </Button>
                            </Group>
                            <Text size="xs" c="dimmed">
                                The microphone streams continuously in 2-second HLS segments, so listeners stay a
                                few seconds behind. Microphone access needs HTTPS or localhost.
                            </Text>
                        </Stack>
                    )}
                </Paper>

                <Paper p="md" radius="md" className={styles.transparent_paper}>
                    <Group justify="space-between" mb="xs">
                        <Title order={4}>Audio quality</Title>
                        <Badge variant="light">converted tracks</Badge>
                    </Group>
                    <Group gap="xs" align="flex-end">
                        <Select
                            label="Target bitrate"
                            description="New uploads, mixes and conversions use this rate. Live voice always streams at 36 kbps for low latency."
                            data={[
                                { value: "24", label: "24 kbps" },
                                { value: "32", label: "32 kbps" },
                                { value: "48", label: "48 kbps (stereo default)" },
                                { value: "64", label: "64 kbps" },
                                { value: "96", label: "96 kbps" },
                                { value: "128", label: "128 kbps" },
                                { value: "160", label: "160 kbps" },
                                { value: "192", label: "192 kbps" },
                                { value: "224", label: "224 kbps" },
                                { value: "256", label: "256 kbps (high)" },
                            ]}
                            value={audioBitRate}
                            onChange={(v) => v && setAudioBitRate(v)}
                            allowDeselect={false}
                            style={{ flex: 1 }}
                        />
                        <Button loading={bitRateBusy} onClick={saveBitRate}>
                            Save
                        </Button>
                    </Group>
                </Paper>

                <SimpleGrid cols={{ base: 1, lg: 2 }} spacing="md">
                    <CartWall />
                    <Stack gap="md">
                        <VoiceOverPanel />
                        <VoiceOverQueue voiceOvers={voiceOvers} />
                    </Stack>
                </SimpleGrid>

                <SimpleGrid cols={{ base: 1, lg: 2 }} spacing="md">
                    <Paper p="md" radius="md" className={styles.transparent_paper}>
                        <Group justify="space-between" mb="xs">
                            <Title order={4}>Coming up</Title>
                            <Text size="xs" c="dimmed">
                                {nextInQueue.length ? "Next in the program log" : "Queue is running out"}
                            </Text>
                        </Group>
                        <Stack gap={6}>
                            {nextInQueue.map((t, i) => (
                                <Flex key={`${t.id}-${i}`} align="center" gap="xs">
                                    <Text size="sm" c="dimmed" w={18}>
                                        {i + 1}
                                    </Text>
                                    <Text size="sm" lineClamp={1} style={{ flex: 1 }}>
                                        {t.name}
                                    </Text>
                                    <KindBadge kind={t.kind} />
                                    <Text size="xs" c="dimmed">
                                        {formatTime(t.duration)}
                                    </Text>
                                </Flex>
                            ))}
                            {!nextInQueue.length ? (
                                <Text size="sm" c="dimmed">
                                    Add music from the Library to keep the station running.
                                </Text>
                            ) : null}
                        </Stack>
                    </Paper>

                    <Paper p="md" radius="md" className={styles.transparent_paper}>
                        <Group justify="space-between" mb="xs">
                            <Title order={4}>Automation</Title>
                            <Badge variant="light">{rules.filter((r) => r.enabled).length} active rules</Badge>
                        </Group>
                        <Stack gap={8}>
                            {rules.slice(0, 5).map((rule) => (
                                <Group key={rule.id} justify="space-between" wrap="nowrap">
                                    <Group gap="xs" wrap="nowrap">
                                        <KindBadge kind={rule.kind} />
                                        <Text size="sm" lineClamp={1}>
                                            {rule.name}
                                        </Text>
                                    </Group>
                                    <Text size="xs" c="dimmed" style={{ flexShrink: 0 }}>
                                        {describeRule(rule.everyNTracks, rule.atMinute)}
                                    </Text>
                                </Group>
                            ))}
                            {!rules.length ? (
                                <Text size="sm" c="dimmed">
                                    No rotation rules yet. Open the Schedule tab to add jingles, ads or station IDs
                                    automatically.
                                </Text>
                            ) : null}
                        </Stack>
                    </Paper>
                </SimpleGrid>
            </Stack>
        </Box>
    );
};

const LiveWaveform: FC<{ analyser: AnalyserNode | null; live: boolean; muted: boolean }> = ({
    analyser,
    live,
    muted,
}) => {
    const canvasRef = useRef<HTMLCanvasElement | null>(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;

        const ctx = canvas.getContext("2d");
        if (!ctx) return;

        let frame = 0;
        let samples: Uint8Array | null = null;

        const draw = () => {
            frame = requestAnimationFrame(draw);

            const dpr = window.devicePixelRatio || 1;
            const width = Math.max(1, Math.floor(canvas.clientWidth * dpr));
            const height = Math.max(1, Math.floor(canvas.clientHeight * dpr));
            if (canvas.width !== width || canvas.height !== height) {
                canvas.width = width;
                canvas.height = height;
            }

            ctx.clearRect(0, 0, width, height);
            const middle = height / 2;
            ctx.lineWidth = Math.max(1, 2 * dpr);
            ctx.strokeStyle = muted ? "rgba(160, 160, 160, 0.45)" : "#fa5252";
            ctx.beginPath();

            if (live && analyser && !muted) {
                if (!samples || samples.length !== analyser.fftSize) {
                    samples = new Uint8Array(analyser.fftSize);
                }
                analyser.getByteTimeDomainData(samples as Uint8Array<ArrayBuffer>);

                const step = samples.length / width;
                for (let x = 0; x < width; x++) {
                    const value = samples[Math.min(samples.length - 1, Math.floor(x * step))] / 128 - 1;
                    const y = middle + value * middle * 0.9;
                    if (x === 0) ctx.moveTo(x, y);
                    else ctx.lineTo(x, y);
                }
            } else {
                ctx.moveTo(0, middle);
                ctx.lineTo(width, middle);
            }

            ctx.stroke();
        };

        frame = requestAnimationFrame(draw);
        return () => cancelAnimationFrame(frame);
    }, [analyser, live, muted]);

    return (
        <Box
            p={4}
            style={{
                borderRadius: 8,
                background: "rgba(0, 0, 0, 0.28)",
                border: "1px solid rgba(255, 255, 255, 0.08)",
            }}
        >
            <canvas ref={canvasRef} style={{ width: "100%", height: 72, display: "block" }} />
        </Box>
    );
};

const CartWall: FC = () => {
    const carts = useCartsStore((s) => s.carts);
    const fetchCarts = useCartsStore((s) => s.fetchCarts);
    const playCart = useCartsStore((s) => s.playCart);
    const deleteCart = useCartsStore((s) => s.deleteCart);
    const [busy, setBusy] = useState("");

    const fire = async (id: string, mode: "next" | "now") => {
        setBusy(`${id}-${mode}`);
        try {
            const { message } = await playCart(id, mode);
            okNotify(message);
        } catch (error) {
            errNotify(error);
        } finally {
            setBusy("");
        }
    };

    const remove = async (id: string) => {
        setBusy(`${id}-delete`);
        try {
            const { message } = await deleteCart(id);
            okNotify(message);
        } catch (error) {
            errNotify(error);
        } finally {
            setBusy("");
        }
    };

    const confirmDelete = (cart: { id: string; name: string }) => {
        modals.openConfirmModal({
            title: "Delete cart audio",
            centered: true,
            children: (
                <Text size="sm">
                    Delete <b>{cart.name}</b> from the library? It will also be removed from the queue and any
                    playlists. This cannot be undone.
                </Text>
            ),
            labels: { confirm: "Delete", cancel: "Cancel" },
            confirmProps: { color: "red" },
            onConfirm: () => remove(cart.id),
        });
    };

    return (
        <Paper p="md" radius="md" className={styles.transparent_paper}>
            <Group justify="space-between" mb="xs">
                <Title order={4}>Cart wall</Title>
                <Tooltip label="Reload carts">
                    <ActionIcon variant="transparent" onClick={() => void fetchCarts()}>
                        <IconReload size={18} color="gray" />
                    </ActionIcon>
                </Tooltip>
            </Group>

            {carts.length ? (
                <SimpleGrid cols={{ base: 2, sm: 3 }} spacing="xs">
                    {carts.map((cart) => (
                        <Card key={cart.id} p="xs" radius="md" withBorder bg="rgba(0,0,0,0.18)">
                            <Group justify="space-between" gap={4} mb={4}>
                                <KindBadge kind={cart.kind} />
                                <Group gap={2}>
                                    <Text size="xs" c="dimmed">
                                        {formatTime(cart.duration)}
                                    </Text>
                                    <Tooltip label="Delete cart audio" openDelay={400}>
                                        <ActionIcon
                                            size="sm"
                                            color="red"
                                            variant="transparent"
                                            loading={busy === `${cart.id}-delete`}
                                            onClick={() => confirmDelete(cart)}
                                        >
                                            <IconTrash size={14} />
                                        </ActionIcon>
                                    </Tooltip>
                                </Group>
                            </Group>
                            <Text size="sm" fw={600} lineClamp={2} mb={6} style={{ minHeight: 36 }}>
                                {cart.name}
                            </Text>
                            <Group gap={4} grow>
                                <Tooltip label="Queue after the current track" openDelay={400}>
                                    <Button
                                        size="compact-xs"
                                        variant="light"
                                        leftSection={<IconQueue size={12} />}
                                        loading={busy === `${cart.id}-next`}
                                        onClick={() => fire(cart.id, "next")}
                                    >
                                        Next
                                    </Button>
                                </Tooltip>
                                <Tooltip label="Play immediately (restarts the current track)" openDelay={400}>
                                    <Button
                                        size="compact-xs"
                                        variant="light"
                                        color="red"
                                        leftSection={<IconPlayerPlayFilled size={12} />}
                                        loading={busy === `${cart.id}-now`}
                                        onClick={() => fire(cart.id, "now")}
                                    >
                                        Now
                                    </Button>
                                </Tooltip>
                            </Group>
                        </Card>
                    ))}
                </SimpleGrid>
            ) : (
                <Text size="sm" c="dimmed">
                    No carts yet. Upload jingles, ads, station IDs or voice links from the Library and they will
                    appear here.
                </Text>
            )}
        </Paper>
    );
};

const VoiceOverPanel: FC = () => {
    const carts = useCartsStore((s) => s.carts);
    const createVoiceOver = useVoiceOverStore((s) => s.createVoiceOver);
    const fetchVoiceOvers = useVoiceOverStore((s) => s.fetchVoiceOvers);
    const fetchPlayback = usePlaybackStore((s) => s.fetchPlayback);
    const isPlaying = usePlaybackStore((s) => s.playback.isPlaying);
    const voiceAssets = carts.filter((c) => c.kind === "voice");

    const [voiceTrackId, setVoiceTrackId] = useState<string | null>(null);
    const [target, setTarget] = useState<"next" | "current">("next");
    const [startOffset, setStartOffset] = useState<number | string>(0);
    const [fadeIn, setFadeIn] = useState<number | string>(1.5);
    const [fadeOut, setFadeOut] = useState<number | string>(2);
    const [duckDb, setDuckDb] = useState<string | null>("-12");
    const [busy, setBusy] = useState(false);

    const submit = async () => {
        if (!voiceTrackId) {
            warnNotify("Choose a voice clip first");
            return;
        }

        setBusy(true);
        try {
            const created = await createVoiceOver({
                voiceTrackId,
                target,
                fadeIn: Number(fadeIn) || 0,
                fadeOut: Number(fadeOut) || 0,
                duckDb: Number(duckDb) || 0,
                startOffset: Number(startOffset) || 0,
            });
            okNotify(
                target === "current"
                    ? `Voice link rendered — ${created.voiceName ?? "voice"} is now talking over the current track`
                    : `Voice link queued over the next track`,
            );
            await Promise.all([fetchPlayback(), fetchVoiceOvers()]);
        } catch (error) {
            errNotify(error);
        } finally {
            setBusy(false);
        }
    };

    return (
        <Paper p="md" radius="md" className={styles.transparent_paper}>
            <Group justify="space-between" mb="xs">
                <Title order={4}>Voice over</Title>
                <Badge variant="light" color="green">
                    {voiceAssets.length ? `${voiceAssets.length} voice clip(s)` : "no voice clips"}
                </Badge>
            </Group>

            {voiceAssets.length ? (
                <Stack gap="xs">
                    <Select
                        label="Voice clip"
                        placeholder="Pick a recorded link"
                        data={voiceAssets.map((v) => ({ value: v.id, label: `${v.name} (${formatTime(v.duration)})` }))}
                        value={voiceTrackId}
                        onChange={setVoiceTrackId}
                        searchable
                    />

                    <SegmentedControl
                        fullWidth
                        value={target}
                        onChange={(value) => setTarget(value as "next" | "current")}
                        data={[
                            { value: "next", label: "Next track" },
                            { value: "current", label: "Current track" },
                        ]}
                    />
                    <Text size="xs" c="dimmed">
                        {target === "current"
                            ? "The current track restarts from the top so the link starts exactly where you specify."
                            : "The link is mixed into the upcoming track before it goes on air."}
                        {!isPlaying && target === "current" ? " Start playback first." : ""}
                    </Text>

                    <SimpleGrid cols={{ base: 2, sm: 4 }} spacing="xs">
                        <NumberInput
                            label="Start at"
                            description="seconds into song"
                            min={0}
                            step={0.5}
                            value={startOffset}
                            onChange={setStartOffset}
                        />
                        <NumberInput
                            label="Voice fade in"
                            description="seconds"
                            min={0}
                            step={0.5}
                            value={fadeIn}
                            onChange={setFadeIn}
                        />
                        <NumberInput
                            label="Voice fade out"
                            description="seconds"
                            min={0}
                            step={0.5}
                            value={fadeOut}
                            onChange={setFadeOut}
                        />
                        <Select
                            label="Duck music"
                            data={DUCK_OPTIONS}
                            value={duckDb}
                            onChange={setDuckDb}
                            allowDeselect={false}
                        />
                    </SimpleGrid>

                    <Button color="green" loading={busy} onClick={submit} disabled={target === "current" && !isPlaying}>
                        {target === "current" ? "Talk over current track" : "Mix into next track"}
                    </Button>
                </Stack>
            ) : (
                <Text size="sm" c="dimmed">
                    Record or upload a voice clip with the Library's “voice” upload option. The RJ's fade in, fade out
                    and ducking values are rendered into the music bed before it is streamed.
                </Text>
            )}
        </Paper>
    );
};

const VoiceOverQueue: FC<{ voiceOvers: VoiceOver[] }> = ({ voiceOvers }) => {
    const cancelVoiceOver = useVoiceOverStore((s) => s.cancelVoiceOver);
    const fetchPlayback = usePlaybackStore((s) => s.fetchPlayback);
    const [busy, setBusy] = useState("");

    const cancel = async (id: string) => {
        setBusy(id);
        try {
            await cancelVoiceOver(id);
            okNotify("Voice link cancelled");
            await fetchPlayback();
        } catch (error) {
            errNotify(error);
        } finally {
            setBusy("");
        }
    };

    const recent = voiceOvers.slice(0, 6);

    return (
        <Paper p="md" radius="md" className={styles.transparent_paper}>
            <Title order={4} mb="xs">
                Link history
            </Title>
            <Stack gap={6}>
                {recent.map((vo) => (
                    <Flex key={vo.id} align="center" gap="xs">
                        <Badge size="xs" variant="light" color={statusColor(vo.status)}>
                            {vo.status}
                        </Badge>
                        <Text size="sm" lineClamp={1} style={{ flex: 1 }}>
                            {vo.voiceName ?? "voice"}
                        </Text>
                        <Text size="xs" c="dimmed">
                            {vo.target === "current" ? "current" : "next"} • duck {vo.duckDb} dB
                        </Text>
                        {vo.status === "pending" || vo.status === "rendered" ? (
                            <Button
                                size="compact-xs"
                                variant="subtle"
                                color="red"
                                loading={busy === vo.id}
                                onClick={() => cancel(vo.id)}
                            >
                                Cancel
                            </Button>
                        ) : null}
                    </Flex>
                ))}
                {!recent.length ? (
                    <Text size="sm" c="dimmed">
                        No voice links yet. A rendered link appears here until it has aired.
                    </Text>
                ) : null}
            </Stack>
        </Paper>
    );
};

const statusColor = (status: VoiceOver["status"]) => {
    switch (status) {
        case "rendered":
            return "green";
        case "played":
            return "gray";
        case "failed":
            return "red";
        case "cancelled":
            return "yellow";
        default:
            return "blue";
    }
};

const describeRule = (everyNTracks: number, atMinute: number) => {
    const parts: string[] = [];
    if (everyNTracks > 0) parts.push(`every ${everyNTracks} track${everyNTracks > 1 ? "s" : ""}`);
    if (atMinute >= 0) parts.push(`at :${String(atMinute).padStart(2, "0")}`);
    return parts.join(" • ") || "no trigger";
};

export const kindOptions = [
    { value: "jingle", label: "Jingle" },
    { value: "ad", label: "Ad" },
    { value: "station_id", label: "Station ID" },
];
