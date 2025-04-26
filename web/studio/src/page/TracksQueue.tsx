import {
    Box,
    Button,
    CloseButton,
    Flex,
    Group,
    LoadingOverlay,
    Paper,
    Space,
    Text,
    useMantineColorScheme,
} from "@mantine/core";
import { DragDropContext, Draggable, Droppable } from "@hello-pangea/dnd";
import { FC, useEffect } from "react";
import { usePlaybackStore } from "../store/playback";
import { useTrackQueueStore } from "../store/track-queue";
import { EmptyLabel } from "../components/EmptyLabel";
import { errNotify, okNotify } from "../notifications";
import { useDisclosure } from "@mantine/hooks";
import { moveArrayItem } from "../utils/array";
import { Track } from "../api/types";
import { handleErr } from "../utils/error";
import { modals } from "@mantine/modals";

export const TrackQueue: FC<{ isMobile?: boolean }> = (props) => {
    const [loader, handLoader] = useDisclosure(false);
    const playback = usePlaybackStore((s) => s.playback);
    const queue = useTrackQueueStore((s) => s.queue);
    const fetchQueue = useTrackQueueStore((s) => s.fetchQueue);
    const updateQueue = useTrackQueueStore((s) => s.updateQueue);
    const removeFromQueue = useTrackQueueStore((s) => s.removeFromQueue);
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

    const handleRemove = async (trackIDs: string[]) => {
        handLoader.open();
        try {
            const { message } = await removeFromQueue(trackIDs);
            okNotify(message);
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
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
            title: "Confirm clear queue",
            centered: true,
            children: <Text size="sm">Do you really want to completely clear the track queue?</Text>,
            labels: { confirm: "Confirm", cancel: "Cancel" },
            onConfirm: () => handleClear(),
        });
    };

    const tracklist = queue.map((track, index) => {
        if (track.id === playback?.currentTrack?.id && playback.isPlaying) return null;

        return (
            <Draggable key={track.id} index={index} draggableId={track.id}>
                {(provided) => (
                    <div {...provided.draggableProps} {...provided.dragHandleProps} ref={provided.innerRef}>
                        <QueueItem track={track} handleRemove={handleRemove} />
                    </div>
                )}
            </Draggable>
        );
    });

    useEffect(() => {
        loadQueue();
    }, []);

    return (
        <Paper radius="md" pos="relative" bg={colorScheme === "dark" ? "dark" : "#f7f7f7"}>
            <Flex p="sm" direction="column" h={props.isMobile ? "calc(100vh - 60px)" : "75vh"} mah={1200}>
                <LoadingOverlay visible={loader} zIndex={1000} />
                <Flex justify="space-between" align="center">
                    <Flex align="center" gap="xs">
                        <Box w={10} h={10} bg={playback?.isPlaying ? "red" : "gray"} style={{ borderRadius: "50%" }} />
                        <Text fw={700} size="lg">
                            Live queue
                        </Text>
                    </Flex>
                    <Text c="dimmed">{queue.length > 1 ? queue.length - (playback.isPlaying ? 1 : 0) : ""}</Text>
                </Flex>

                <Space h={12} />

                <Box flex={1} style={{ overflow: "auto", overflowX: "hidden" }}>
                    <DragDropContext
                        onDragEnd={async ({ destination, source }) => {
                            try {
                                await updateQueue(moveArrayItem(queue, source.index, destination?.index || 0));
                            } catch (error) {
                                handleErr(error);
                            }
                        }}
                    >
                        <Droppable droppableId="dnd-list" direction="vertical">
                            {(provided) => (
                                <Flex direction="column" {...provided.droppableProps} ref={provided.innerRef}>
                                    {tracklist}
                                    {provided.placeholder}
                                </Flex>
                            )}
                        </Droppable>
                    </DragDropContext>
                    {!queue.length || (queue.length === 1 && playback.isPlaying) ? (
                        <EmptyLabel label={"Queue is empty"} />
                    ) : null}
                </Box>

                <Space h={12} />

                <Group gap="xs">
                    <Button onClick={confirmClear} variant="light" color="red" disabled={loader || queue.length <= 1}>
                        Clear
                    </Button>
                </Group>
            </Flex>
        </Paper>
    );
};

const QueueItem: FC<{ track: Track; handleRemove: (ids: string[]) => Promise<void> }> = ({ track, handleRemove }) => {
    return (
        <Paper p="xs" key={track.id} mb="xs">
            <Flex justify="space-between" align="center">
                <Text style={{ whiteSpace: "nowrap", textOverflow: "ellipsis", overflow: "hidden" }}>{track.name}</Text>
                <CloseButton size="sm" onClick={() => handleRemove([track.id])} />
            </Flex>
        </Paper>
    );
};
