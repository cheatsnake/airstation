import {
    ActionIcon,
    Box,
    Flex,
    MantineSize,
    Paper,
    Progress,
    Space,
    Text,
    Tooltip,
    useMantineTheme,
} from "@mantine/core";
import { FC, useEffect, useRef, useState } from "react";
import { airstationAPI, API_HOST } from "../api";
import { usePlaybackStore } from "../store/playback";
import { formatTime } from "../utils/time";
import { useTrackQueueStore } from "../store/track-queue";
import { IconHeadphones, IconPlayerPlayFilled, IconPlayerStopFilled, IconVolumeOff, IconVolumeOn } from "../icons";
import { useDisclosure } from "@mantine/hooks";
import { errNotify } from "../notifications";
import { EVENTS, useEventSourceStore } from "../store/events";
import { PlaybackState } from "../api/types";
import { SettingsModal } from "./Settings";
import Hls from "hls.js";
import { modals } from "@mantine/modals";
import styles from "./styles.module.css";

export const Playback: FC<{ isMobile?: boolean }> = (props) => {
    const updateIntervalID = useRef(0);
    const [loader, handLoader] = useDisclosure(false);
    const playback = usePlaybackStore((s) => s.playback);
    const setPlayback = usePlaybackStore((s) => s.setPlayback);
    const fetchPlayback = usePlaybackStore((s) => s.fetchPlayback);
    const syncElapsedTime = usePlaybackStore((s) => s.syncElapsedTime);
    const rotateQueue = useTrackQueueStore((s) => s.rotateQueue);
    const addEventHandler = useEventSourceStore((s) => s.addEventHandler);
    const theme = useMantineTheme();

    useEffect(() => {
        (async () => {
            await fetchPlayback();
        })();

        addEventHandler(EVENTS.newTrack, async () => {
            rotateQueue();
            await fetchPlayback();
        });

        if (!updateIntervalID.current) {
            updateIntervalID.current = setInterval(() => {
                syncElapsedTime();
            }, 1000);
        }

        return () => {
            if (updateIntervalID.current) {
                clearInterval(updateIntervalID.current);
                updateIntervalID.current = 0;
            }
        };
    }, []);

    const togglePlayback = async () => {
        handLoader.open();
        try {
            const pb = playback.isPlaying ? await airstationAPI.pausePlayback() : await airstationAPI.playPlayback();
            setPlayback(pb);
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
        }
    };

    const handlePlaybackAction = () => {
        if (!playback.isPlaying) {
            togglePlayback();
            return;
        }

        modals.openConfirmModal({
            title: "Confirm stop playback",
            cancelProps: { variant: "light", color: "gray" },
            centered: true,
            children: (
                <Text size="sm">
                    Do you really want to stop playing tracks on the station? This action will affect all listeners.
                </Text>
            ),
            labels: { confirm: "Confirm", cancel: "Cancel" },
            onConfirm: () => togglePlayback(),
        });
    };

    return (
        <Paper w="100%" radius="md" className={styles.transparent_paper}>
            <Flex
                p="sm"
                gap="md"
                w="100%"
                h={props.isMobile ? "calc(100vh - 60px)" : undefined}
                direction={props.isMobile ? "column-reverse" : "row"}
                justify={props.isMobile ? "space-between" : "cener"}
                align="center"
            >
                {props.isMobile ? <Box h="1" /> : null}
                <Flex gap={props.isMobile ? "lg" : "xs"} direction="column" justify="center" align="center">
                    <Tooltip openDelay={500} label={`${playback.isPlaying ? "Stop" : "Start"} playback of the stream`}>
                        <ActionIcon
                            onClick={handlePlaybackAction}
                            disabled={loader}
                            color="dark"
                            variant="subtle"
                            size={props.isMobile ? 100 : "md"}
                            aria-label="Settings"
                        >
                            {playback?.isPlaying ? (
                                <IconPlayerStopFilled stroke="0" fill={theme.colors[theme.primaryColor][8]} />
                            ) : (
                                <IconPlayerPlayFilled stroke="0" fill={theme.colors.dark[8]} />
                            )}
                        </ActionIcon>
                    </Tooltip>
                    <StreamToggler size={props.isMobile ? "xl" : "md"} playback={playback} />
                </Flex>
                <Flex w="100%" direction="column">
                    <Flex justify="space-between" direction={props.isMobile ? "column-reverse" : "row"} gap="sm">
                        {playback.isPlaying ? (
                            <Text>{playback?.currentTrack?.name}</Text>
                        ) : (
                            <Text c="dimmed">Stream is stopped</Text>
                        )}
                        <SettingsModal />
                    </Flex>
                    <Space h={10} />
                    <Box>
                        <Progress
                            radius="xl"
                            value={(playback.currentTrackElapsed / (playback?.currentTrack?.duration || 1)) * 100}
                        />
                        <Flex justify="space-between" align="center" mt={4}>
                            <Text ta="end" mt={3} c="dimmed">
                                {`${formatTime(playback?.currentTrackElapsed || 0)} / ${formatTime(playback?.currentTrack?.duration || 0)}`}
                            </Text>
                            <ListenerCounter />
                        </Flex>
                    </Box>
                </Flex>
            </Flex>
        </Paper>
    );
};

const ListenerCounter = () => {
    const [count, setCount] = useState(0);
    const addEventHandler = useEventSourceStore((s) => s.addEventHandler);

    const handleCounter = (msg: MessageEvent<string>) => {
        setCount(Number(msg.data));
    };

    useEffect(() => {
        addEventHandler(EVENTS.countListeners, handleCounter);
    }, []);

    return (
        <Tooltip openDelay={500} label={`Listener counter`}>
            <Flex gap={5} justify="center" align="center" opacity={0.5}>
                <IconHeadphones size={18} />
                <Text>{!count ? "" : count}</Text>
            </Flex>
        </Tooltip>
    );
};

const StreamToggler: FC<{ playback: PlaybackState; size: MantineSize }> = (props) => {
    const videoRef = useRef<HTMLAudioElement | null>(null);
    const streamRef = useRef<Hls | null>(null);
    const [isPlaying, setIsPlaying] = useState(false);

    const initStream = () => {
        if (isPlaying) return;

        streamRef.current = new Hls();
        streamRef.current.loadSource(API_HOST + "/stream");
        streamRef.current.attachMedia(videoRef.current as unknown as HTMLMediaElement);
    };

    const destroyStream = () => {
        streamRef.current?.destroy();
        streamRef.current = null;
        setIsPlaying(false);
    };

    const handlePause = () => {
        videoRef.current?.pause();
        destroyStream();
    };

    const handlePlay = async () => {
        try {
            initStream();
            await videoRef.current?.play();
            setIsPlaying(true);
        } catch (error) {
            console.log("Failed to play: ", error);
        }
    };

    useEffect(() => {
        if (!props.playback.isPlaying && streamRef.current) {
            destroyStream();
        }
    }, [props.playback]);

    return (
        <>
            <audio
                style={{ display: "none" }}
                onPause={handlePause}
                onPlay={handlePlay}
                id="stream"
                ref={videoRef}
            ></audio>
            <Tooltip openDelay={500} label={`${isPlaying ? "Mute" : "Unmute"} the stream just for me`}>
                <ActionIcon
                    variant="subtle"
                    onClick={isPlaying ? () => videoRef.current?.pause() : () => videoRef.current?.play()}
                    color="gray"
                    size={props.size}
                    aria-label="Settings"
                >
                    {isPlaying ? <IconVolumeOff size={20} /> : <IconVolumeOn size={20} />}
                </ActionIcon>
            </Tooltip>
        </>
    );
};
