import { Box, Flex, Paper, Space, Text } from "@mantine/core";
import { FC, useEffect } from "react";
import { airstationAPI } from "../api";
import { usePlaybackStore } from "../store/playback";
import { useTrackQueueStore } from "../store/track-queue";
import { EmptyLabel } from "../components/EmptyLabel";
import { errNotify } from "../notifications";

export const TrackQueue: FC<{}> = () => {
    const playback = usePlaybackStore((s) => s.playback);
    const queue = useTrackQueueStore((s) => s.queue);
    const setQueue = useTrackQueueStore((s) => s.setQueue);

    const loadQueue = async () => {
        try {
            const q = await airstationAPI.getQueue();
            setQueue(q);
        } catch (error) {
            errNotify(error);
        }
    };

    useEffect(() => {
        loadQueue();
    }, []);

    return (
        <Paper withBorder p="sm" pos="relative">
            <Flex justify="space-between" align="center">
                <Flex align="center" gap="xs">
                    <Box w={10} h={10} bg={playback?.isPlaying ? "red" : "gray"} style={{ borderRadius: "50%" }} />
                    <Text fw={700} size="lg">
                        Live queue
                    </Text>
                </Flex>
                <Text c="dimmed">{`${queue.length - 1} ${queue.length > 2 ? "tracks" : "track"}`}</Text>
            </Flex>
            <Space h={12} />

            <Flex direction="column" gap="sm" mih={100}>
                {queue.length > 1 ? (
                    queue
                        .filter((t) => playback?.currentTrack.id != t.id)
                        .map((track) => (
                            <Paper p="xs" withBorder key={track.id}>
                                {track.name}
                            </Paper>
                        ))
                ) : (
                    <EmptyLabel label={"Queue is empty"} />
                )}
            </Flex>
        </Paper>
    );
};
