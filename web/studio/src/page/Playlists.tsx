import { FC, useEffect, useState } from "react";
import { useDebouncedState, useDisclosure } from "@mantine/hooks";
import {
    Modal,
    Flex,
    Box,
    Button,
    Tooltip,
    ActionIcon,
    TextInput,
    Textarea,
    Text,
    LoadingOverlay,
    CloseButton,
    Select,
} from "@mantine/core";
import { modals } from "@mantine/modals";
import { CSS } from "@dnd-kit/utilities";
import { DndContext, DragEndEvent } from "@dnd-kit/core";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";
import { SortableContext, useSortable } from "@dnd-kit/sortable";
import { airstationAPI } from "../api";
import { formatTime } from "../utils/time";
import { Playlist, Track } from "../api/types";
import { moveArrayItem } from "../utils/array";
import { usePlaybackStore } from "../store/playback";
import { usePlaylistStore } from "../store/playlists";
import { EmptyLabel } from "../components/EmptyLabel";
import { useTrackQueueStore } from "../store/track-queue";
import { errNotify, okNotify, warnNotify } from "../notifications";
import { IconPlayerPlayFilled, IconPlaylist, IconTrash } from "../icons";

const PlaylistItem: FC<{ data: Playlist; confirmDelete: (id: string) => void }> = ({ data, confirmDelete }) => {
    const [hovered, setHovered] = useState(false);
    const updateQueue = useTrackQueueStore((s) => s.updateQueue);
    const pausePlayback = usePlaybackStore((s) => s.pause);
    const playPlayback = usePlaybackStore((s) => s.play);

    const handleUse = async () => {
        try {
            const p = await airstationAPI.getPlaylist(data.id);
            if (!p.tracks.length) {
                warnNotify("Playlist is empty");
                return;
            }

            await pausePlayback();
            await updateQueue(p.tracks);
            await playPlayback();

            okNotify("Playlist added to a queue");
        } catch (error) {
            errNotify(error);
        }
    };

    return (
        <Flex
            align="center"
            justify="space-between"
            gap="sm"
            onMouseEnter={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}
            w="100%"
        >
            <Flex gap="xs" w="100%">
                <PlaylistModal data={data} hovered={hovered} setHovered={setHovered} />
                <Flex gap={5}>
                    <Tooltip label="Delete playlist">
                        <ActionIcon variant="transparent" color="red" onClick={() => confirmDelete(data.id)}>
                            <IconTrash size={18} />
                        </ActionIcon>
                    </Tooltip>
                    <Tooltip label="Put it in a queue">
                        <ActionIcon onClick={handleUse} disabled={!data.trackCount} variant="transparent" color="green">
                            <IconPlayerPlayFilled size={18} />
                        </ActionIcon>
                    </Tooltip>
                </Flex>
            </Flex>
        </Flex>
    );
};

const TrackItem: FC<{
    track: Track;
    index: number;
    setTracks: React.Dispatch<React.SetStateAction<Track[]>>;
}> = ({ index, setTracks, track }) => {
    const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id: track.id });
    const style = { transform: CSS.Transform.toString(transform), transition };

    return (
        <Flex ref={setNodeRef} p="0.3rem" gap="sm" align="start" style={style}>
            <Flex {...attributes} {...listeners} style={{ cursor: "grab" }} w="100%">
                <Text lh="xs" w="100%">
                    {`${index + 1}. ${track.name}`}
                </Text>
                <Text c="dimmed">{formatTime(Math.round(track.duration))}</Text>
            </Flex>
            <CloseButton onClick={() => setTracks((prev) => prev.filter((pt) => pt.id !== track.id))} />
        </Flex>
    );
};

const TrackList: FC<{
    tracks: Track[];
    setTracks: React.Dispatch<React.SetStateAction<Track[]>>;
}> = ({ tracks, setTracks }) => {
    const handleDragEvent = (event: DragEndEvent) => {
        const { active, over } = event;
        if (over && active.id !== over.id) {
            setTracks((tracks) => {
                const fromIndex = tracks.findIndex((t) => t.id === active.id);
                const toIndex = tracks.findIndex((t) => t.id === over.id);
                return moveArrayItem(tracks, fromIndex, toIndex);
            });
        }
    };

    return (
        <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={handleDragEvent}>
            <SortableContext items={tracks}>
                {tracks.map((track, index) => (
                    <TrackItem key={track.id} track={track} index={index} setTracks={setTracks} />
                ))}
            </SortableContext>
        </DndContext>
    );
};

const PlaylistModal: FC<{
    data: Playlist;
    hovered: boolean;
    setHovered: React.Dispatch<React.SetStateAction<boolean>>;
}> = ({ data, hovered, setHovered }) => {
    const [opened, { open, close }] = useDisclosure(false);
    const [isLoading, loading] = useDisclosure(false);
    const [tracks, setTracks] = useState<Track[]>([]);
    const [name, setName] = useState(data.name);
    const [descr, setDescr] = useState(data.description);
    const editPlaylist = usePlaylistStore((s) => s.editPlaylist);

    const loadTracks = async () => {
        loading.open();
        try {
            const p = await airstationAPI.getPlaylist(data.id);
            setTracks(p.tracks || []);
        } catch (error) {
            errNotify(error);
        } finally {
            loading.close();
        }
    };

    const handleSave = async () => {
        loading.open();
        try {
            const { message } = await editPlaylist(
                data.id,
                name,
                tracks.map((t) => t.id),
                descr,
            );
            okNotify(message);
            close();
        } catch (error) {
            errNotify(error);
        } finally {
            loading.close();
        }
    };

    const handleOpen = () => {
        open();
        setHovered(false);
    };

    useEffect(() => {
        if (opened) loadTracks();
    }, [opened]);

    return (
        <>
            <Modal size="lg" centered opened={opened} onClose={close} withCloseButton={false} stackId={`playlist-${data.id}`}>
                <LoadingOverlay visible={isLoading} overlayProps={{ radius: "md" }} />
                <Box>
                    <TextInput
                        required
                        minLength={3}
                        placeholder="Name*"
                        value={name}
                        onChange={(event) => setName(event.currentTarget.value)}
                    />
                    <Textarea
                        mt="sm"
                        placeholder="Description"
                        maxLength={4096}
                        minRows={1}
                        autosize
                        maxRows={4}
                        value={descr}
                        onChange={(event) => setDescr(event.currentTarget.value)}
                    />
                    <Box
                        mt="md"
                        flex={1}
                        mih={200}
                        mah={500}
                        style={{
                            overflowX: "hidden",
                            overflowY: "auto",
                            scrollbarGutter: "stable",
                        }}
                    >
                        <TrackList tracks={tracks} setTracks={setTracks} />
                        {!tracks.length ? <EmptyLabel label="No tracks" /> : null}
                    </Box>
                </Box>

                <Flex mt="md" justify="space-between" align="center" gap="sm">
                    <Flex justify="end">
                        <Text>Total time:&nbsp;</Text>
                        <Text c="dimmed">
                            {formatTime(Math.round(tracks.reduce((prev, curr) => (prev += curr.duration), 0)))}
                        </Text>
                    </Flex>
                    <Flex gap="sm">
                        <Button autoFocus onClick={close} color="dimmed" variant="light">
                            Close
                        </Button>
                        <Button onClick={handleSave} color="green" variant="light">
                            Save
                        </Button>
                    </Flex>
                </Flex>
            </Modal>
            <Flex w="100%" gap="sm" onClick={handleOpen} style={{ cursor: "pointer" }}>
                <Text style={{ textWrap: "nowrap" }} c={hovered ? "blue" : undefined}>
                    {data.name}
                </Text>
                <Text w="100%" c="dimmed">
                    {data.trackCount} {data.trackCount === 1 ? "track" : "tracks"}
                </Text>
            </Flex>
        </>
    );
};

const CreatePlaylistModal: FC<{ opened: boolean; onClose: () => void }> = ({ opened, onClose }) => {
    const [isLoading, loading] = useDisclosure(false);
    const [name, setName] = useState("");
    const [descr, setDescr] = useState("");
    const addPlaylist = usePlaylistStore((s) => s.addPlaylist);

    const handleClose = () => {
        setName("");
        setDescr("");
        onClose();
    };

    const handleCreate = async () => {
        loading.open();
        try {
            await addPlaylist(name.trim(), [], descr);
            okNotify("A new playlist has been successfully created.");
            handleClose();
        } catch (error) {
            errNotify(error);
        } finally {
            loading.close();
        }
    };

    return (
        <Modal centered opened={opened} onClose={handleClose} title="New playlist" stackId="create-playlist">
            <Flex
                direction="column"
                gap="md"
                component="form"
                onSubmit={(event) => {
                    event.preventDefault();
                    if (name.trim().length < 3) return;
                    void handleCreate();
                }}
            >
                <TextInput
                    required
                    autoFocus
                    placeholder="Name*"
                    value={name}
                    onChange={(event) => setName(event.currentTarget.value)}
                />
                <Textarea
                    placeholder="Description"
                    maxLength={4096}
                    rows={4}
                    value={descr}
                    onChange={(event) => setDescr(event.currentTarget.value)}
                />
                <Flex justify="flex-end" gap="sm">
                    <Button onClick={handleClose} variant="light" color="dimmed">
                        Cancel
                    </Button>
                    <Button type="submit" loading={isLoading} variant="light" disabled={name.trim().length < 3}>
                        Create
                    </Button>
                </Flex>
            </Flex>
        </Modal>
    );
};

export const PlaylistsModal: FC<{}> = () => {
    const [opened, { open, close }] = useDisclosure(false);
    const [createOpened, { open: openCreate, close: closeCreate }] = useDisclosure(false);
    const playlists = usePlaylistStore((s) => s.playlists);
    const fetchPlaylists = usePlaylistStore((s) => s.fetchPlaylists);
    const deletePlaylist = usePlaylistStore((s) => s.deletePlaylist);
    const [search, setSearch] = useDebouncedState("", 200);
    const [isLoading, loading] = useDisclosure(true);

    const closeAll = () => {
        closeCreate();
        close();
    };

    const loadPlaylists = async () => {
        loading.open();
        try {
            await fetchPlaylists();
        } catch (error) {
            errNotify(error);
        } finally {
            loading.close();
        }
    };

    const handleDelete = async (id: string) => {
        loading.open();
        try {
            const { message } = await deletePlaylist(id);
            okNotify(message);
        } catch (error) {
            errNotify(error);
        } finally {
            loading.close();
        }
    };

    const confirmDelete = (id: string) => {
        modals.openConfirmModal({
            title: "Confirm delete playlist",
            cancelProps: { variant: "light", color: "gray" },
            centered: true,
            children: (
                <Text size="sm">
                    Do you really want to delete this playlist? All tracks from the playlist will still be available in
                    the library.
                </Text>
            ),
            labels: { confirm: "Confirm", cancel: "Cancel" },
            onConfirm: () => handleDelete(id),
        });
    };

    useEffect(() => {
        loadPlaylists();
    }, []);

    return (
        <Modal.Stack>
            <Modal
                centered
                size="lg"
                opened={opened}
                onClose={closeAll}
                withCloseButton={false}
                radius="md"
                stackId="playlists"
            >
                <Flex direction="column" gap="md">
                    <Text>Playlists</Text>
                    {playlists.length > 7 ? (
                        <TextInput
                            defaultValue={search}
                            onChange={(event) => setSearch(event.currentTarget.value)}
                            placeholder="Search..."
                        />
                    ) : null}
                    <Box flex={1} mih={200} mah="90vh" style={{ overflowY: "auto" }}>
                        <LoadingOverlay visible={isLoading} overlayProps={{ radius: "md", opacity: 0.7 }} />
                        {playlists
                            .filter((p) => p.name.toLowerCase().includes(search))
                            .map((p) => (
                                <PlaylistItem key={p.id} data={p} confirmDelete={confirmDelete} />
                            ))}
                        {!playlists.filter((p) => p.name.toLowerCase().includes(search)).length ? (
                            <EmptyLabel label="No playlists" />
                        ) : null}
                    </Box>
                    <Flex justify="end" gap="sm">
                        <Button onClick={closeAll} color="dimmed" variant="light">
                            Close
                        </Button>
                        <Button onClick={openCreate} variant="light" color="green">
                            New playlist
                        </Button>
                    </Flex>
                </Flex>
            </Modal>

            <CreatePlaylistModal opened={createOpened} onClose={closeCreate} />

            <Tooltip openDelay={500} label="Playlists">
                <ActionIcon onClick={open} variant="transparent" size="md">
                    <IconPlaylist size={18} color="gray" />
                </ActionIcon>
            </Tooltip>
        </Modal.Stack>
    );
};

export const AddToPalylistModal: FC<{ trackIDs: string[]; opened: boolean; onClose: () => void }> = ({
    trackIDs,
    opened,
    onClose,
}) => {
    const playlists = usePlaylistStore((s) => s.playlists);
    const [selected, setSelected] = useState<string | null>(null);
    const [newName, setNewName] = useState("");
    const fetchPlaylists = usePlaylistStore((s) => s.fetchPlaylists);
    const editPlaylist = usePlaylistStore((s) => s.editPlaylist);
    const addPlaylist = usePlaylistStore((s) => s.addPlaylist);
    const [isLoading, loading] = useDisclosure(true);

    const handleClose = () => {
        setSelected(null);
        setNewName("");
        onClose();
    };

    const loadPlaylists = async () => {
        loading.open();
        try {
            await fetchPlaylists();
        } catch (error) {
            errNotify(error);
        } finally {
            loading.close();
        }
    };

    const handleConfirm = async () => {
        if (!selected && newName.trim().length < 3) return;
        loading.open();
        try {
            if (selected) {
                const p = await airstationAPI.getPlaylist(selected);
                const { message } = await editPlaylist(
                    p.id,
                    p.name,
                    [...new Set([...(p.tracks || [])?.map((t) => t.id), ...trackIDs])],
                    p.description,
                );
                okNotify(message);
            } else {
                await addPlaylist(newName.trim(), trackIDs);
                okNotify("Playlist created with the selected tracks.");
            }
            handleClose();
        } catch (error) {
            errNotify(error);
        } finally {
            loading.close();
        }
    };

    useEffect(() => {
        if (opened) loadPlaylists();
    }, [opened]);

    return (
        <Modal size="sm" centered onClose={handleClose} opened={opened} withCloseButton={false}>
            <LoadingOverlay visible={isLoading} overlayProps={{ radius: "md" }} />
            <Select
                searchable
                label="Add to an existing playlist"
                placeholder={playlists.length ? "Select a playlist" : "No playlists yet"}
                value={selected}
                onChange={setSelected}
                data={[{ group: "", items: playlists.map((p) => ({ label: p.name, value: p.id })) }]}
            />
            <TextInput
                mt="md"
                label="Or create a new playlist"
                placeholder="New playlist name"
                value={newName}
                onChange={(event) => {
                    setNewName(event.currentTarget.value);
                    if (event.currentTarget.value) setSelected(null);
                }}
            />
            <Flex mt="md" gap="sm" justify="end">
                <Button onClick={handleClose} color="dimmed" variant="light">
                    Cancel
                </Button>
                <Button
                    disabled={!selected && newName.trim().length < 3}
                    variant="light"
                    onClick={handleConfirm}
                >
                    Confirm
                </Button>
            </Flex>
        </Modal>
    );
};
