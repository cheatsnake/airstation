import { Flex, LoadingOverlay, Paper, Space, Text, TextInput } from "@mantine/core";
import { FC, useEffect, useState } from "react";
import { airstationAPI } from "../api";
import { useTrackQueueStore } from "../store/track-queue";
import { useTracksStore } from "../store/tracks";
import { EmptyLabel } from "../components/EmptyLabel";
import { useDebouncedValue, useDisclosure } from "@mantine/hooks";

export const TrackLibrary: FC<{}> = () => {
    const [search, setSearch] = useState("");
    const [debouncedSearch] = useDebouncedValue(search, 500);
    const [loader, handLoader] = useDisclosure(false);

    const tracks = useTracksStore((s) => s.tracks);
    const queue = useTrackQueueStore((s) => s.queue);
    const setTracks = useTracksStore((s) => s.setTracks);

    const loadTracks = async (page = 1, limit = 100) => {
        try {
            handLoader.open();
            const result = await airstationAPI.getTracks(page, limit, search);
            setTracks(result.tracks);
        } catch (error) {
            console.log(error);
        } finally {
            handLoader.close();
        }
    };

    const isTrackInQueue = (trackID: string) => {
        return queue.map((t) => t.id).includes(trackID);
    };

    useEffect(() => {
        loadTracks();
    }, [debouncedSearch]);

    return (
        <Paper withBorder p="xs" pos="relative">
            <LoadingOverlay visible={loader} zIndex={1000} />

            <Flex justify="space-between" align="center">
                <Text fw={700} size="lg">
                    Tracks library
                </Text>
                <Text c="dimmed">{`${tracks.length} ${tracks.length > 1 ? "tracks" : "track"}`}</Text>
            </Flex>
            <Space h={12} />
            <TextInput placeholder="Search" value={search} onChange={(event) => setSearch(event.currentTarget.value)} />
            <Space h={16} />
            <Flex direction="column" gap="sm" mih={100} justify="center">
                {tracks.length ? (
                    tracks.map((track) => (
                        <Paper p="xs" withBorder key={track.id} c={isTrackInQueue(track.id) ? "dimmed" : undefined}>
                            {track.name}
                        </Paper>
                    ))
                ) : (
                    <EmptyLabel label={"No tracks found"} />
                )}
            </Flex>
        </Paper>
    );
};
