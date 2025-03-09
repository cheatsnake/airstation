import { Box, CloseButton, Flex, LoadingOverlay, Paper, Space, Text, useMantineColorScheme } from "@mantine/core";
import { FC, useEffect } from "react";
import { usePlaybackStore } from "../store/playback";
import { useTrackQueueStore } from "../store/track-queue";
import { EmptyLabel } from "../components/EmptyLabel";
import { errNotify, okNotify } from "../notifications";
import { useDisclosure } from "@mantine/hooks";
import { airstationAPI } from "../api";

export const TrackQueue: FC<{}> = () => {
    const [loader, handLoader] = useDisclosure(false);
    const playback = usePlaybackStore((s) => s.playback);
    const queue = useTrackQueueStore((s) => s.queue);
    const fetchQueue = useTrackQueueStore((s) => s.fetchQueue);
    const { colorScheme } = useMantineColorScheme();

    const loadQueue = async () => {
        handLoader.open();
        try {
            await fetchQueue();
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
        }
    };

    const handleRemove = async (trackID: string) => {
        handLoader.open();
        try {
            const { message } = await airstationAPI.removeFromQueue([trackID]);
            await fetchQueue();
            okNotify(message);
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
        }
    };

    useEffect(() => {
        loadQueue();
    }, []);

    return (
        <Paper p="sm" radius="md" pos="relative" bg={colorScheme === "dark" ? "dark" : "#f7f7f7"}>
            <LoadingOverlay visible={loader} zIndex={1000} />
            <Flex justify="space-between" align="center">
                <Flex align="center" gap="xs">
                    <Box w={10} h={10} bg={playback?.isPlaying ? "red" : "gray"} style={{ borderRadius: "50%" }} />
                    <Text fw={700} size="lg">
                        Live queue
                    </Text>
                </Flex>
                <Text c="dimmed">{`${queue.length > 0 ? queue.length - 1 : 0} ${
                    queue.length > 2 ? "tracks" : "track"
                }`}</Text>
            </Flex>

            <Space h={12} />

            <Flex direction="column" gap="sm" mih={100}>
                {queue.length > 1 ? (
                    queue
                        .filter((t) => playback?.currentTrack?.id != t.id)
                        .map((track) => (
                            <Paper p="xs" withBorder key={track.id}>
                                <Flex justify="space-between" align="center">
                                    <Text
                                        style={{ whiteSpace: "nowrap", textOverflow: "ellipsis", overflow: "hidden" }}
                                    >
                                        {track.name}
                                    </Text>
                                    <CloseButton onClick={() => handleRemove(track.id)} />
                                </Flex>
                            </Paper>
                        ))
                ) : (
                    <EmptyLabel label={"Queue is empty"} />
                )}
            </Flex>
        </Paper>
    );
};
