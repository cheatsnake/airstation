import { FC, useCallback, useEffect, useRef, useState } from "react";
import {
    ActionIcon,
    Box,
    Button,
    Checkbox,
    FileButton,
    Flex,
    Group,
    LoadingOverlay,
    Menu,
    Modal,
    Paper,
    Select,
    Space,
    Stack,
    Text,
    TextInput,
    Tooltip,
} from "@mantine/core";
import { modals } from "@mantine/modals";
import { useDebouncedValue, useDisclosure } from "@mantine/hooks";
import { AddToPalylistModal } from "./Playlists";
import { TRACK_KINDS, Track, TrackKind } from "../api/types";
import { DisclosureHandler } from "../types";
import { useTracksStore } from "../store/tracks";
import { EmptyLabel } from "../components/EmptyLabel";
import { AudioPlayer } from "../components/AudioPlayer";
import { useTrackQueueStore } from "../store/track-queue";
import { EVENTS, useEventSourceStore } from "../store/events";
import { errNotify, infoNotify, okNotify, warnNotify } from "../notifications";
import { IconPlaylist, IconQueue, IconReload, IconSortAscending, IconSortDescending, IconTrash } from "../icons";
import styles from "./styles.module.css";

const PAGE_LIMIT = 20;

const KIND_FILTERS = [{ value: "all", label: "All assets" }, ...TRACK_KINDS.map((k) => ({ value: k.value, label: k.label }))];

export const TrackLibrary: FC<{ isMobile?: boolean }> = (props) => {
    const tracksContainerRef = useRef<HTMLDivElement>(null);
    const [page, setPage] = useState(1);
    const [search, setSearch] = useState("");
    const [kind, setKind] = useState<TrackKind | "all">("all");
    const [debouncedSearch] = useDebouncedValue(search, 500);
    const [sortBy, setSortBy] = useState<keyof Track>("id");
    const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
    const [playingTrackID, setPlayingTrackID] = useState("");
    const [selectedTrackIDs, setSelectedTrackIDs] = useState<Set<string>>(new Set());
    const [loader, handLoader] = useDisclosure(false);
    const [uploadOpened, upload] = useDisclosure(false);
    const addEventHandler = useEventSourceStore((s) => s.addEventHandler);
    const [hovered, setHovered] = useState(false);

    const tracks = useTracksStore((s) => s.tracks);
    const totalTracks = useTracksStore((s) => s.totalTracks);
    const queue = useTrackQueueStore((s) => s.queue);
    const fetchTracks = useTracksStore((s) => s.fetchTracks);

    const loadTracks = useCallback(async () => {
        handLoader.open();
        try {
            await fetchTracks(page, PAGE_LIMIT, search, sortBy, sortOrder, kind);
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
        }
    }, [page, search, sortBy, sortOrder, kind]);

    const toggleTrackPlaying = (id: string) => {
        // If the same track is clicked again, pause it
        // If a different track is clicked, pause the current one and play the new one
        const newVelue = playingTrackID === id ? "" : id;
        setPlayingTrackID(newVelue);
    };

    const isTrackInQueue = (trackID: string) => {
        return queue.map((t) => t.id).includes(trackID);
    };

    const handleSort = (sb: keyof Track, so: "asc" | "desc") => {
        setSortBy(sb);
        setSortOrder(so);
    };

    const handleLoadNextPage = () => {
        setPage((prev) => prev + 1);
    };

    useEffect(() => {
        addEventHandler(EVENTS.loadedTracks, (msg: MessageEvent<string>) => {
            setSortBy("id");
            setSortOrder("desc");
            setSearch("");
            setPage(1);
            loadTracks();

            infoNotify(`${msg.data} new track(s) are now available in your library.`);
        });
    }, []);

    useEffect(() => {
        setPage(1);
    }, [debouncedSearch, sortBy, sortOrder, kind]);

    useEffect(() => {
        loadTracks();
    }, [page, debouncedSearch, sortBy, sortOrder, kind]);

    return (
        <Paper radius="md" className={styles.transparent_paper} h="100%">
            <Flex p="xs" direction="column" h={props.isMobile ? "calc(100vh - 70px)" : "100%"} mah={1200}>
                <LoadingOverlay visible={loader} overlayProps={{ radius: "md" }} />
                <UploadModal opened={uploadOpened} close={upload.close} handLoader={handLoader} />

                <Flex justify="space-between" align="center">
                    <Flex align="center" gap="xs">
                        <Text fw={700} size="lg">
                            Library
                        </Text>
                        <Text c="dimmed">{`${tracks.length}/${totalTracks} ${
                            totalTracks > 1 ? "tracks" : "track"
                        }`}</Text>
                    </Flex>

                    <Flex align="center" gap={5}>
                        <Tooltip openDelay={500} label="Reload list">
                            <ActionIcon onClick={loadTracks} variant="transparent" size="md">
                                <IconReload size={18} color="gray" />
                            </ActionIcon>
                        </Tooltip>
                        <Tooltip openDelay={500} label={`Sort by ${sortOrder === "asc" ? "descending" : "ascending"}`}>
                            <ActionIcon
                                onClick={() => handleSort(sortBy, sortOrder === "asc" ? "desc" : "asc")}
                                variant="transparent"
                                size="md"
                            >
                                {sortOrder === "asc" ? (
                                    <IconSortAscending size={18} color="gray" />
                                ) : (
                                    <IconSortDescending size={18} color="gray" />
                                )}
                            </ActionIcon>
                        </Tooltip>
                        <Tooltip openDelay={500} label="Parameter by which tracks are sorted">
                            <Select
                                w={90}
                                withCheckIcon={false}
                                variant="transparent"
                                size="xs"
                                allowDeselect={false}
                                value={sortBy}
                                data={["id", "name", "duration"]}
                                onChange={(value) => handleSort(value as keyof Track, sortOrder)}
                                comboboxProps={{ offset: 0, variant: "transparent" }}
                            />
                        </Tooltip>
                    </Flex>
                </Flex>

                <Space h={12} />

                <Flex gap="xs" wrap="wrap">
                    <TextInput
                        style={{ flexGrow: 1, minWidth: 150 }}
                        placeholder="Search"
                        value={search}
                        onChange={(event) => setSearch(event.currentTarget.value)}
                    />
                    <Select
                        w={props.isMobile ? "100%" : 140}
                        data={KIND_FILTERS}
                        value={kind}
                        onChange={(value) => setKind((value as TrackKind | "all") ?? "all")}
                        allowDeselect={false}
                        aria-label="Filter by asset kind"
                    />
                    <Button variant="light" color="green" onClick={upload.open}>
                        Add
                    </Button>
                </Flex>

                <Space h={16} />

                <Box
                    flex={1}
                    onMouseEnter={() => setHovered(true)}
                    onMouseLeave={() => setHovered(false)}
                    style={{
                        overflowX: "hidden",
                        overflowY: hovered ? "auto" : "hidden",
                        scrollbarGutter: "stable",
                    }}
                    ref={tracksContainerRef}
                >
                    <Flex direction="column" gap="sm" justify="center">
                        {tracks.length ? (
                            <>
                                {tracks.map((track) => (
                                    <Paper
                                        p="xs"
                                        key={track.id}
                                        c={isTrackInQueue(track.id) ? "dimmed" : undefined}
                                        bg="transparent"
                                    >
                                        <AudioPlayer
                                            track={track}
                                            isPlaying={playingTrackID === track.id}
                                            selected={selectedTrackIDs}
                                            setSelected={setSelectedTrackIDs}
                                            togglePlaying={() => toggleTrackPlaying(track.id)}
                                        />
                                    </Paper>
                                ))}
                                {Math.ceil(totalTracks / PAGE_LIMIT) > page ? (
                                    <Button onClick={handleLoadNextPage} variant="transparent">
                                        Load more
                                    </Button>
                                ) : null}
                            </>
                        ) : (
                            <EmptyLabel label={"No tracks found"} />
                        )}
                    </Flex>
                </Box>
                <Space h={12} />

                <Group justify="space-between">
                    {selectedTrackIDs.size ? <Text>{`Selected: ${selectedTrackIDs.size}`}</Text> : <div />}
                    <TrackActions
                        handLoader={handLoader}
                        selected={selectedTrackIDs}
                        setSelected={setSelectedTrackIDs}
                        tracks={tracks}
                    />
                </Group>
            </Flex>
        </Paper>
    );
};

const UploadModal: FC<{ opened: boolean; close: () => void; handLoader: DisclosureHandler }> = (props) => {
    const uploadTracks = useTracksStore((s) => s.uploadTracks);
    const [kind, setKind] = useState<TrackKind>("music");

    const handleUpload = async (files: File[]) => {
        if (!files.length) {
            warnNotify("No files for upload...");
            return;
        }

        props.handLoader.open();
        try {
            const { message } = await uploadTracks(files, kind);
            okNotify(message);
            props.close();
        } catch (error) {
            errNotify(error);
        } finally {
            props.handLoader.close();
        }
    };

    const description =
        kind === "music"
            ? "Songs for rotation. Music tracks must be at least 15 seconds long."
            : kind === "voice"
              ? "Recorded RJ links. They get default fade in/out and ducking values that the console can override."
              : "Short station assets. Jingles, ads and IDs can be as short as one second.";

    return (
        <Modal opened={props.opened} onClose={props.close} title="Upload assets" centered>
            <Stack gap="sm">
                <Select
                    label="Asset kind"
                    data={TRACK_KINDS.map((k) => ({ value: k.value, label: k.label }))}
                    value={kind}
                    onChange={(value) => setKind((value as TrackKind) ?? "music")}
                    allowDeselect={false}
                />
                <Text size="xs" c="dimmed">
                    {description}
                </Text>
                <FileButton multiple onChange={handleUpload} accept="audio/mpeg,audio/aac,audio/wav,audio/flac,audio/mp4">
                    {(fileProps) => (
                        <Button {...fileProps} variant="light" color="green">
                            Choose files & upload
                        </Button>
                    )}
                </FileButton>
            </Stack>
        </Modal>
    );
};

const TrackActions: FC<{
    handLoader: DisclosureHandler;
    selected: Set<string>;
    setSelected: React.Dispatch<React.SetStateAction<Set<string>>>;
    tracks: Track[];
}> = (props) => {
    const addToQueue = useTrackQueueStore((s) => s.addToQueue);
    const deleteTracks = useTracksStore((s) => s.deleteTracks);
    const [playlistOpened, playlist] = useDisclosure(false);

    const toggleSelection = () => {
        if (props.selected.size) {
            props.setSelected(new Set());
            return;
        }

        props.setSelected(new Set(props.tracks.map((t) => t.id)));
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
            await addToQueue([...props.selected]);
            props.setSelected(new Set());
        } catch (error) {
            errNotify(error);
        } finally {
            props.handLoader.close();
        }
    };

    const confirmDelete = () => {
        modals.openConfirmModal({
            title: "Confirm clear queue",
            cancelProps: { variant: "light", color: "gray" },
            centered: true,
            children: (
                <Text size="sm">
                    Do you really want to delete the selected tracks from the server? They will also be removed from
                    the queue and any playlists.
                </Text>
            ),
            labels: { confirm: "Confirm", cancel: "Cancel" },
            onConfirm: () => handleDelete(),
        });
    };

    return (
        <Flex align="center" gap="md">
            <Menu position="left" offset={0} keepMounted>
                <Menu.Target>
                    <Button variant="light" px="md" disabled={!props.selected.size}>
                        Action
                    </Button>
                </Menu.Target>

                <Menu.Dropdown>
                    <Menu.Item leftSection={<IconQueue size={14} />} color="blue" onClick={handleAddToQueue}>
                        Move to queue
                    </Menu.Item>
                    <Menu.Divider />
                    <Menu.Item leftSection={<IconPlaylist size={14} />} onClick={playlist.open}>
                        Add to playlist
                    </Menu.Item>
                    <Menu.Divider />
                    <Menu.Item leftSection={<IconTrash size={14} />} color="red" onClick={confirmDelete}>
                        Delete
                    </Menu.Item>
                </Menu.Dropdown>
            </Menu>
            <AddToPalylistModal trackIDs={[...props.selected]} opened={playlistOpened} onClose={playlist.close} />
            <Checkbox size="md" readOnly checked={props.selected.size > 0} onClick={toggleSelection} />
        </Flex>
    );
};
