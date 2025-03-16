import { Box, Flex, Paper, Progress, Space, Text } from "@mantine/core";
import { FC, useEffect, useRef } from "react";
import { API_HOST, API_PREFIX } from "../api";
import { usePlaybackStore } from "../store/playback";
import { formatTime } from "../utils/time";
import { useTrackQueueStore } from "../store/track-queue";

const EVENTS = {
    newTrack: "new_track",
};

const EVENT_SOURCE_URL = API_HOST + API_PREFIX + "/events";

export const Playback: FC<{}> = () => {
    const intervalID = useRef(0);
    const playback = usePlaybackStore((s) => s.playback);
    const fetchPlayback = usePlaybackStore((s) => s.fetchPlayback);
    const incElapsedTime = usePlaybackStore((s) => s.incElapsedTime);
    const rotateQueue = useTrackQueueStore((s) => s.rotateQueue);

    useEffect(() => {
        const events = new EventSource(EVENT_SOURCE_URL);

        (async () => {
            await fetchPlayback();
            intervalID.current = setInterval(() => {
                incElapsedTime(1);
            }, 1000);
        })();

        events.addEventListener(EVENTS.newTrack, async () => {
            rotateQueue();
            await fetchPlayback();
        });

        return () => clearInterval(intervalID.current);
    }, []);

    return (
        <Paper p="sm" w="100%" h={95} pos="relative">
            <Flex gap="sm">
                <Box w="100%">
                    <Text>{playback?.currentTrack?.name || "Unknown"}</Text>
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
