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
import { airstationAPI } from "../api";
import { moveArrayItem } from "../utils/array";
import { Track } from "../api/types";
import { handleErr } from "../utils/error";

export const TrackQueue: FC<{}> = () => {
    const [loader, handLoader] = useDisclosure(false);
    const playback = usePlaybackStore((s) => s.playback);
    const queue = useTrackQueueStore((s) => s.queue);
    const fetchQueue = useTrackQueueStore((s) => s.fetchQueue);
    const updateQueue = useTrackQueueStore((s) => s.updateQueue);
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

    const tracklist = queue.map((track, index) => {
        if (track.id === playback?.currentTrack?.id) return null;

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

            <Box h="100%" mah={650} style={{ overflow: "auto", overflowX: "hidden" }}>
                {queue.length > 1 ? (
                    <DragDropContext
                        onDragEnd={async ({ destination, source }) => {
                            handLoader.open();
                            try {
                                await updateQueue(moveArrayItem(queue, source.index, destination?.index || 0));
                            } catch (error) {
                                handleErr(error);
                            } finally {
                                handLoader.close();
                            }
                        }}
                    >
                        <Droppable droppableId="dnd-list" direction="vertical">
                            {(provided) => (
                                <Flex direction="column" mih={100} {...provided.droppableProps} ref={provided.innerRef}>
                                    {tracklist}
                                    {provided.placeholder}
                                </Flex>
                            )}
                        </Droppable>
                    </DragDropContext>
                ) : (
                    <EmptyLabel label={"Queue is empty"} />
                )}
            </Box>

            <Space h={12} />

            <Group gap="xs">
                {playback?.isPlaying ? (
                    <Button variant="light" color="yellow" disabled>
                        Pause
                    </Button>
                ) : (
                    <Button variant="light" color="green">
                        Play
                    </Button>
                )}
                <Button variant="light" color="red" disabled>
                    Clear
                </Button>
            </Group>
        </Paper>
    );
};

const QueueItem: FC<{ track: Track; handleRemove: (id: string) => Promise<void> }> = ({ track, handleRemove }) => {
    return (
        <Paper p="xs" key={track.id} mb="xs">
            <Flex justify="space-between" align="center">
                <Text style={{ whiteSpace: "nowrap", textOverflow: "ellipsis", overflow: "hidden" }}>{track.name}</Text>
                <CloseButton size="sm" onClick={() => handleRemove(track.id)} />
            </Flex>
        </Paper>
    );
};
