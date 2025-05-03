import { ActionIcon, Box, Flex, MantineSize, Paper, Progress, Space, Text, Tooltip } from "@mantine/core";
import { FC, useEffect, useRef, useState } from "react";
import { airstationAPI, API_HOST } from "../api";
import { usePlaybackStore } from "../store/playback";
import { formatTime } from "../utils/time";
import { useTrackQueueStore } from "../store/track-queue";
import { IconHeadphones, IconPlayerPlayFilled, IconPlayerStopFilled, IconVolumeOff, IconVolumeOn } from "../icons";
import { useDisclosure } from "@mantine/hooks";
import { errNotify } from "../notifications";
import { EVENTS, useEventSourceStore } from "../store/events";
import { useThemeBlackColor } from "../hooks/useThemeBlackColor";
import { PlaybackState } from "../api/types";
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

    useEffect(() => {
        (async () => {
            await fetchPlayback();
        })();

        addEventHandler(EVENTS.newTrack, async () => {
            rotateQueue();
            await fetchPlayback();
        });

        return () => {
            clearInterval(updateIntervalID.current);
            updateIntervalID.current = 0;
        };
    }, []);

    useEffect(() => {
        if (playback.isPlaying && updateIntervalID.current === 0) {
            updateIntervalID.current = setInterval(() => {
                syncElapsedTime();
            }, 1000);
        }

        if (!playback.isPlaying && updateIntervalID.current !== 0) {
            clearInterval(updateIntervalID.current);
            updateIntervalID.current = 0;
        }
    }, [playback]);

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
        <Paper
            w="100%"
            radius="md"
            // bg={colorScheme === "dark" ? "dark" : "#f7f7f7"}
            style={{
                backgroundColor: "rgb(0 0 0 / 30%)",
                backdropFilter: "blur(10px)",
            }}
        >
            <Flex
                p="sm"
                gap="md"
                w="100%"
                h={props.isMobile ? "calc(100vh - 60px)" : "95"}
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
                            variant="subtle"
                            color={useThemeBlackColor()}
                            size={props.isMobile ? 100 : "md"}
                            aria-label="Settings"
                        >
                            {playback?.isPlaying ? (
                                <IconPlayerStopFilled fill={useThemeBlackColor()} />
                            ) : (
                                <IconPlayerPlayFilled fill={useThemeBlackColor()} />
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
                        <ListenerCounter />
                    </Flex>
                    <Space h={10} />
                    <Box>
                        <Progress
                            className={styles.progress_bar}
                            radius="xl"
                            value={(playback.currentTrackElapsed / (playback?.currentTrack?.duration || 1)) * 100}
                        />
                        <Text ta="end" mt={3} c="dimmed">
                            {formatTime(playback?.currentTrackElapsed || 0)}/
                            {formatTime(playback?.currentTrack?.duration || 0)}
                        </Text>
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
    const videoRef = useRef<HTMLVideoElement | null>(null);
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

    const togglePlayback = async () => {
        try {
            initStream();

            if (isPlaying) {
                videoRef.current?.pause();
                destroyStream();
                setIsPlaying(false);
            } else {
                await videoRef.current?.play();
                setIsPlaying(true);
            }
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
            <video style={{ display: "none" }} id="stream" ref={videoRef}></video>
            <Tooltip openDelay={500} label={`${isPlaying ? "Mute" : "Unmute"} the stream just for me`}>
                <ActionIcon
                    onClick={togglePlayback}
                    variant="subtle"
                    color={useThemeBlackColor()}
                    size={props.size}
                    aria-label="Settings"
                >
                    {isPlaying ? (
                        <IconVolumeOn fill={useThemeBlackColor()} />
                    ) : (
                        <IconVolumeOff fill={useThemeBlackColor()} />
                    )}
                </ActionIcon>
            </Tooltip>
        </>
    );
};
