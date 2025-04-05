import { ActionIcon, Box, Flex, Paper, Progress, Space, Text } from "@mantine/core";
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

export const Playback: FC<{}> = () => {
    const updateIntervalID = useRef(0);
    const [loader, handLoader] = useDisclosure(false);
    const playback = usePlaybackStore((s) => s.playback);
    const setPlayback = usePlaybackStore((s) => s.setPlayback);
    const fetchPlayback = usePlaybackStore((s) => s.fetchPlayback);
    const incElapsedTime = usePlaybackStore((s) => s.incElapsedTime);
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

        return () => clearInterval(updateIntervalID.current);
    }, []);

    useEffect(() => {
        if (playback.isPlaying && updateIntervalID.current === 0) {
            updateIntervalID.current = setInterval(() => {
                incElapsedTime(1);
            }, 1000);
        }

        if (!playback.isPlaying && updateIntervalID.current !== 0) {
            clearInterval(updateIntervalID.current);
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

    return (
        <Paper p="sm" w="100%" h={95}>
            <Flex gap="sm" justify="center" align="center">
                <Flex gap="xs">
                    <ActionIcon
                        onClick={togglePlayback}
                        disabled={loader}
                        variant="subtle"
                        color={useThemeBlackColor()}
                        size="sm"
                        aria-label="Settings"
                    >
                        {playback?.isPlaying ? (
                            <IconPlayerStopFilled fill={useThemeBlackColor()} />
                        ) : (
                            <IconPlayerPlayFilled fill={useThemeBlackColor()} />
                        )}
                    </ActionIcon>
                    <StreamToggler playback={playback} />
                </Flex>
                <Box w="100%">
                    <Flex justify="space-between" align="center">
                        <Text>{playback?.currentTrack?.name || "Unknown"}</Text>
                        <ListenersCounter />
                    </Flex>
                    <Space h={10} />
                    <Progress
                        color="air"
                        radius="xl"
                        value={(playback.currentTrackElapsed / (playback?.currentTrack?.duration || 1)) * 100}
                    />
                    <Text ta="end" mt={3} c="dimmed">
                        {formatTime(playback?.currentTrackElapsed || 0)}/
                        {formatTime(playback?.currentTrack?.duration || 0)}
                    </Text>
                </Box>
            </Flex>
        </Paper>
    );
};

const ListenersCounter = () => {
    const [count, setCount] = useState(0);
    const addEventHandler = useEventSourceStore((s) => s.addEventHandler);

    const handleCounter = (msg: MessageEvent<string>) => {
        setCount(Number(msg.data));
    };

    useEffect(() => {
        addEventHandler(EVENTS.countListeners, handleCounter);
    }, []);

    return (
        <Flex gap={5} justify="center" align="center" opacity={0.5}>
            <IconHeadphones size={18} />
            <Text>{count}</Text>
        </Flex>
    );
};

const StreamToggler: FC<{ playback: PlaybackState }> = (props) => {
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
            <ActionIcon
                onClick={togglePlayback}
                variant="subtle"
                color={useThemeBlackColor()}
                size="sm"
                aria-label="Settings"
            >
                {isPlaying ? (
                    <IconVolumeOn fill={useThemeBlackColor()} />
                ) : (
                    <IconVolumeOff fill={useThemeBlackColor()} />
                )}
            </ActionIcon>
        </>
    );
};
