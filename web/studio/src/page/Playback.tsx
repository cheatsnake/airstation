import { ActionIcon, Box, Flex, Paper, Progress, Space, Text } from "@mantine/core";
import { FC, useEffect, useRef, useState } from "react";
import { airstationAPI } from "../api";
import { usePlaybackStore } from "../store/playback";
import { formatTime } from "../utils/time";
import { useTrackQueueStore } from "../store/track-queue";
import { IconHeadphones, IconPlayerPlayFilled, IconPlayerStopFilled } from "../icons";
import { useDisclosure } from "@mantine/hooks";
import { errNotify } from "../notifications";
import { EVENTS, useEventSourceStore } from "../store/events";

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
            updateIntervalID.current = setInterval(() => {
                incElapsedTime(1);
            }, 1000);
        })();

        addEventHandler(EVENTS.newTrack, async () => {
            rotateQueue();
            await fetchPlayback();
        });

        return () => clearInterval(updateIntervalID.current);
    }, []);

    const togglePlayback = async () => {
        handLoader.open();
        try {
            const pb = playback.isPlaying ? await airstationAPI.pausePlayback() : await airstationAPI.playPlayback();

            if (pb.isPlaying) {
                updateIntervalID.current = setInterval(() => {
                    incElapsedTime(1);
                }, 1000);
            } else {
                clearInterval(updateIntervalID.current);
            }

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
                <Box>
                    <ActionIcon
                        onClick={togglePlayback}
                        disabled={loader}
                        variant="subtle"
                        color="black"
                        size="sm"
                        aria-label="Settings"
                    >
                        {playback?.isPlaying ? (
                            <IconPlayerStopFilled fill="black" />
                        ) : (
                            <IconPlayerPlayFilled fill="black" />
                        )}
                    </ActionIcon>
                </Box>
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
