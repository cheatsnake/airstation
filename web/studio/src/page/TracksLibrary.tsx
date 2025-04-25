import {
    Box,
    Button,
    Checkbox,
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
import { useTrackQueueStore } from "../store/track-queue";
import { useTracksStore } from "../store/tracks";
import { EmptyLabel } from "../components/EmptyLabel";
import { useDebouncedValue, useDisclosure } from "@mantine/hooks";
import { AudioPlayer } from "../components/AudioPlayer";
import { errNotify, infoNotify, okNotify, warnNotify } from "../notifications";
import { DisclosureHandler } from "../types";
import { Track } from "../api/types";
import { modals } from "@mantine/modals";
import { EVENTS, useEventSourceStore } from "../store/events";

export const TrackLibrary: FC<{ isMobile?: boolean }> = (props) => {
    const [search, setSearch] = useState("");
    const [playingTrackID, setPlayingTrackID] = useState("");
    const [selectedTrackIDs, setSelectedTrackIDs] = useState<Set<string>>(new Set());
    const [debouncedSearch] = useDebouncedValue(search, 500);
    const [loader, handLoader] = useDisclosure(false);
    const addEventHandler = useEventSourceStore((s) => s.addEventHandler);
    const { colorScheme } = useMantineColorScheme();

    const tracks = useTracksStore((s) => s.tracks);
    const queue = useTrackQueueStore((s) => s.queue);
    const fetchTracks = useTracksStore((s) => s.fetchTracks);

    const loadTracks = async (page = 1, limit = 100) => {
        handLoader.open();
        try {
            await fetchTracks(page, limit, search);
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
        addEventHandler(EVENTS.loadedTracks, async (msg: MessageEvent<string>) => {
            await loadTracks();
            infoNotify(`${msg.data} new track(s) are now available in your library.`);
        });
    }, []);

    useEffect(() => {
        loadTracks();
    }, [debouncedSearch]);

    return (
        <Paper radius="md" pos="relative" bg={colorScheme === "dark" ? "dark" : "#f7f7f7"}>
            <Flex p="xs" direction="column" h={props.isMobile ? "calc(100vh - 60px)" : "75vh"} mah={1200}>
                <LoadingOverlay visible={loader} zIndex={1000} />

                <Flex justify="space-between" align="center">
                    <Text fw={700} size="lg">
                        Tracks library
                    </Text>
                    <Text c="dimmed">{`${tracks.length} ${tracks.length > 1 ? "tracks" : "track"}`}</Text>
                </Flex>

                <Space h={12} />

                <Flex gap="xs">
                    <TextInput
                        style={{ flexGrow: 1 }}
                        placeholder="Search"
                        value={search}
                        onChange={(event) => setSearch(event.currentTarget.value)}
                    />
                    <TrackUploader handLoader={handLoader} />
                </Flex>

                <Space h={16} />

                <Box flex={1} style={{ overflow: "auto", overflowX: "hidden" }}>
                    <Flex direction="column" gap="sm" justify="center">
                        {tracks.length ? (
                            tracks.map((track) => (
                                <Paper p="xs" key={track.id} c={isTrackInQueue(track.id) ? "dimmed" : undefined}>
                                    <AudioPlayer
                                        track={track}
                                        isPlaying={playingTrackID === track.id}
                                        isTrackInQueue={isTrackInQueue(track.id)}
                                        selected={selectedTrackIDs}
                                        setSelected={setSelectedTrackIDs}
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

                <Group justify="space-between">
                    {selectedTrackIDs.size ? <Text c="dimmed">{`Selected: ${selectedTrackIDs.size}`}</Text> : <div />}
                    <TrackActions
                        handLoader={handLoader}
                        selected={selectedTrackIDs}
                        setSelected={setSelectedTrackIDs}
                        availableTracks={tracks.filter((t) => !isTrackInQueue(t.id))}
                        disabled={!selectedTrackIDs.size}
                    />
                </Group>
            </Flex>
        </Paper>
    );
};

const TrackUploader: FC<{ handLoader: DisclosureHandler }> = (props) => {
    const uploadTracks = useTracksStore((s) => s.uploadTracks);

    const handleUpload = async (files: File[]) => {
        if (!files.length) {
            warnNotify("No files for upload...");
            return;
        }

        props.handLoader.open();
        try {
            const { message } = await uploadTracks(files);
            okNotify(message);
        } catch (error) {
            errNotify(error);
        } finally {
            props.handLoader.close();
        }
    };

    return (
        <>
            <FileButton multiple onChange={handleUpload} accept="audio/*">
                {(props) => (
                    <Button {...props} variant="light" color="green">
                        Add
                    </Button>
                )}
            </FileButton>
        </>
    );
};

const TrackActions: FC<{
    handLoader: DisclosureHandler;
    selected: Set<string>;
    setSelected: React.Dispatch<React.SetStateAction<Set<string>>>;
    availableTracks: Track[];
    disabled?: boolean;
}> = (props) => {
    const addToQueue = useTrackQueueStore((s) => s.addToQueue);
    const deleteTracks = useTracksStore((s) => s.deleteTracks);

    const toggleSelection = () => {
        if (props.selected.size) {
            props.setSelected(new Set());
            return;
        }

        props.setSelected(new Set(props.availableTracks.map((t) => t.id)));
    };

    const handleDelete = async () => {
        props.handLoader.open();
        try {
            const { message } = await deleteTracks([...props.selected]);
            props.setSelected(new Set());
            okNotify(message);
        } catch (error) {
            errNotify(error);
        } finally {
            props.handLoader.close();
        }
    };

    const handleAddToQueue = async () => {
        props.handLoader.open();
        try {
            const { message } = await addToQueue([...props.selected]);
            props.setSelected(new Set());
            okNotify(message);
        } catch (error) {
            errNotify(error);
        } finally {
            props.handLoader.close();
        }
    };

    const confirmDelete = () => {
        modals.openConfirmModal({
            title: "Confirm clear queue",
            centered: true,
            children: <Text size="sm">Do you really want to delete selected tracks from the server?</Text>,
            labels: { confirm: "Confirm", cancel: "Cancel" },
            onConfirm: () => handleDelete(),
        });
    };

    return (
        <Flex align="center" gap="xs">
            <Button disabled={props.disabled} onClick={confirmDelete} variant="light" color="red">
                Delete
            </Button>
            <Button disabled={props.disabled} onClick={handleAddToQueue} variant="light">
                Queue
            </Button>
            <Checkbox size="md" color="dimmed" readOnly checked={props.selected.size > 0} onClick={toggleSelection} />
        </Flex>
    );
};
