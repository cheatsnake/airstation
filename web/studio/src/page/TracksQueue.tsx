import { FC, useEffect, useState } from "react";
import {
    ActionIcon,
    Box,
    Button,
    CloseButton,
    Flex,
    Group,
    LoadingOverlay,
    Paper,
    Space,
    Text,
    Tooltip,
} from "@mantine/core";
import { modals } from "@mantine/modals";
import { useDisclosure } from "@mantine/hooks";
import { CSS } from "@dnd-kit/utilities";
import { DndContext, DragEndEvent } from "@dnd-kit/core";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";
import { SortableContext, useSortable } from "@dnd-kit/sortable";
import { PlaylistsModal } from "./Playlists";
import { Track } from "../api/types";
import { IconReload } from "../icons";
import { usePlaybackStore } from "../store/playback";
import { EmptyLabel } from "../components/EmptyLabel";
import { errNotify, okNotify } from "../notifications";
import { useTrackQueueStore } from "../store/track-queue";
import { moveArrayItem, shuffleArray } from "../utils/array";
import styles from "./styles.module.css";

export const TrackQueue: FC<{ isMobile?: boolean }> = (props) => {
    const [loader, handLoader] = useDisclosure(false);
    const playback = usePlaybackStore((s) => s.playback);
    const queue = useTrackQueueStore((s) => s.queue);
    const fetchQueue = useTrackQueueStore((s) => s.fetchQueue);
    const updateQueue = useTrackQueueStore((s) => s.updateQueue);
    const removeFromQueue = useTrackQueueStore((s) => s.removeFromQueue);
    const [hovered, setHovered] = useState(false);

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

    const handleRemove = async (trackIDs: string[]) => {
        handLoader.open();
        try {
            await removeFromQueue(trackIDs);
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
            setHovered(true);
        }
    };

    const handleClear = async () => {
        handLoader.open();
        try {
            const trackIDs = queue.filter(({ id }) => id !== playback.currentTrack?.id).map(({ id }) => id);
            const { message } = await removeFromQueue(trackIDs);
            okNotify(message);
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
        }
    };

    const confirmClear = () => {
        modals.openConfirmModal({
            title: "Confirm clear the queue",
            cancelProps: { variant: "light", color: "gray" },
            centered: true,
            children: <Text size="sm">Do you really want to completely clear the track queue?</Text>,
            labels: { confirm: "Confirm", cancel: "Cancel" },
            onConfirm: () => handleClear(),
        });
    };

    const handleShuffle = async () => {
        try {
            const shuffled = shuffleArray(queue.filter(({ id }) => id !== playback.currentTrack?.id));
            await updateQueue(playback.currentTrack ? [playback.currentTrack, ...shuffled] : shuffled);
        } catch (error) {
            errNotify(error);
        }
    };

    const confirmShuffle = () => {
        modals.openConfirmModal({
            title: "Confirm shuffle the queue",
            cancelProps: { variant: "light", color: "gray" },
            centered: true,
            children: <Text size="sm">Do you really want to shuffle the track queue?</Text>,
            labels: { confirm: "Confirm", cancel: "Cancel" },
            onConfirm: () => handleShuffle(),
        });
    };

    const handleDragEvent = async (event: DragEndEvent) => {
        const { active, over } = event;
        if (over && active.id !== over.id) {
            const fromIndex = queue.findIndex((t) => t.id === active.id);
            const toIndex = queue.findIndex((t) => t.id === over.id);
            setHovered(false);
            try {
                await updateQueue(moveArrayItem(queue, fromIndex, toIndex));
            } catch (error) {
                errNotify(error);
            }
        }
    };

    const tracklist = queue.map((track) => {
        if (track.id === playback?.currentTrack?.id && playback.isPlaying) return null;
        return <QueueItem key={track.id} track={track} handleRemove={handleRemove} />;
    });

    useEffect(() => {
        loadQueue();
    }, []);

    return (
        <Paper radius="md" className={styles.transparent_paper}>
            <Flex p="sm" direction="column" h={props.isMobile ? "calc(100vh - 60px)" : "80vh"} mah={1200}>
                <LoadingOverlay visible={loader} overlayProps={{ radius: "md" }} />
                <Flex justify="space-between" align="center">
                    <Flex align="center" justify="center" gap="xs">
                        <Box
                            w={10}
                            h={10}
                            bg={playback?.isPlaying ? "red" : "#ffffff33"}
                            style={{ borderRadius: "50%" }}
                        />
                        <Text fw={700} size="lg">
                            Live queue
                        </Text>
                        <Text c="dimmed">{queue.length > 1 ? queue.length - (playback.isPlaying ? 1 : 0) : ""}</Text>
                    </Flex>
                    <Flex>
                        <PlaylistsModal />
                        <Tooltip openDelay={500} label="Reload list">
                            <ActionIcon onClick={loadQueue} variant="transparent" size="md">
                                <IconReload size={18} color="gray" />
                            </ActionIcon>
                        </Tooltip>
                    </Flex>
                </Flex>

                <Space h={12} />

                <Box
                    flex={1}
                    onMouseEnter={() => setHovered(true)}
                    onMouseLeave={() => setHovered(false)}
                    style={{
                        overflowX: "hidden",
                        overflowY: hovered ? "auto" : "hidden",
                        scrollbarGutter: "stable",
                    }}
                >
                    <DndContext modifiers={[restrictToVerticalAxis]} onDragEnd={handleDragEvent}>
                        <SortableContext items={queue}>{tracklist}</SortableContext>
                    </DndContext>
                    {!queue.length || (queue.length === 1 && playback.isPlaying) ? (
                        <EmptyLabel label={"Queue is empty"} />
                    ) : null}
                </Box>

                <Space h={12} />

                <Group gap="xs">
                    <Button onClick={confirmClear} variant="light" color="gray" disabled={loader || queue.length <= 1}>
                        Clear
                    </Button>
                    <Button onClick={confirmShuffle} variant="light" color="pink" disabled={queue.length < 3}>
                        🎲 Shuffle
                    </Button>
                </Group>
            </Flex>
        </Paper>
    );
};

const QueueItem: FC<{ track: Track; handleRemove: (ids: string[]) => Promise<void> }> = ({ track, handleRemove }) => {
    const [hovered, setHovered] = useState(false);
    const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id: track.id });
    const style = { transform: CSS.Transform.toString(transform), transition };

    return (
        <Paper
            ref={setNodeRef}
            p="0.3rem"
            key={track.id}
            bg="transparent"
            style={style}
            onMouseEnter={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}
        >
            <Flex align="center" gap={5}>
                <Text
                    {...attributes}
                    {...listeners}
                    w="100%"
                    c={hovered ? "air" : undefined}
                    style={{ whiteSpace: "nowrap", textOverflow: "ellipsis", overflow: "hidden", cursor: "grab" }}
                >
                    {track.name}
                </Text>
                <CloseButton
                    variant="transparent"
                    display={hovered ? undefined : "none"}
                    size="sm"
                    onClick={() => {
                        handleRemove([track.id]);
                        setHovered(false);
                    }}
                />
            </Flex>
        </Paper>
    );
};
