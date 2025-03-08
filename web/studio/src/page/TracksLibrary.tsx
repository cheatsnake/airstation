import {
    Box,
    Button,
    FileButton,
    Flex,
    Group,
    LoadingOverlay,
    Paper,
    Space,
    Text,
    TextInput,
    useMantineColorScheme,
} from "@mantine/core";
import { FC, useEffect, useState } from "react";
import { airstationAPI } from "../api";
import { useTrackQueueStore } from "../store/track-queue";
import { useTracksStore } from "../store/tracks";
import { EmptyLabel } from "../components/EmptyLabel";
import { useDebouncedValue, useDisclosure } from "@mantine/hooks";
import { AudioPlayer } from "../components/AudioPlayer";
import { errNotify, warnNotify } from "../notifications";

export const TrackLibrary: FC<{}> = () => {
    const [search, setSearch] = useState("");
    const [playingTrackID, setPlayingTrackID] = useState("");
    const [debouncedSearch] = useDebouncedValue(search, 500);
    const [loader, handLoader] = useDisclosure(false);
    const { colorScheme } = useMantineColorScheme();

    const tracks = useTracksStore((s) => s.tracks);
    const queue = useTrackQueueStore((s) => s.queue);
    const setTracks = useTracksStore((s) => s.setTracks);

    const loadTracks = async (page = 1, limit = 100) => {
        try {
            handLoader.open();
            const result = await airstationAPI.getTracks(page, limit, search);
            setTracks(result.tracks);
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
        }
    };

    const toggleTrackPlaying = (id: string) => {
        // If the same track is clicked again, pause it
        // If a different track is clicked, pause the current one and play the new one
        const newVelue = playingTrackID === id ? "" : id;
        setPlayingTrackID(newVelue);
    };

    const isTrackInQueue = (trackID: string) => {
        return queue.map((t) => t.id).includes(trackID);
    };

    useEffect(() => {
        loadTracks();
    }, [debouncedSearch]);

    return (
        <Paper p="xs" pos="relative" bg={colorScheme === "dark" ? "dark" : "#f7f7f7"}>
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
            <Box h={600} style={{ overflow: "auto", overflowX: "hidden" }}>
                <Flex direction="column" gap="sm" justify="center">
                    {tracks.length ? (
                        tracks.map((track) => (
                            <Paper p="xs" key={track.id} c={isTrackInQueue(track.id) ? "dimmed" : undefined}>
                                <AudioPlayer
                                    track={track}
                                    isPlaying={playingTrackID === track.id}
                                    togglePlaying={() => toggleTrackPlaying(track.id)}
                                />
                            </Paper>
                        ))
                    ) : (
                        <EmptyLabel label={"No tracks found"} />
                    )}
                </Flex>
            </Box>
            <Space h={12} />
            <TrackUploader />
        </Paper>
    );
};

const TrackUploader = () => {
    const [loader, handLoader] = useDisclosure(false);
    const addTracks = useTracksStore((s) => s.addTracks);

    const handleUpload = async (files: File[]) => {
        if (!files.length) {
            warnNotify("No files for upload...");
            return;
        }

        handLoader.open();
        try {
            const tracks = await airstationAPI.uploadTracks(files);
            addTracks(tracks);
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
        }
    };

    return (
        <>
            <Group justify="flex-end">
                <LoadingOverlay visible={loader} loaderProps={{ type: "dots", size: "lg" }} />
                <FileButton multiple onChange={handleUpload} accept="audio/*">
                    {(props) => (
                        <Button {...props} variant="light">
                            Add
                        </Button>
                    )}
                </FileButton>
            </Group>
        </>
    );
};
