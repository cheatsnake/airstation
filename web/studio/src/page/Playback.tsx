import { Box, Flex, Paper, Progress, Space, Text } from "@mantine/core";
import { FC, useEffect, useState } from "react";
import { airstationAPI } from "../api";
import { usePlaybackStore } from "../store/playback";
import { formatTime } from "../utils/time";
import { errNotify } from "../notifications";

export const Playback: FC<{}> = () => {
    const playback = usePlaybackStore((s) => s.playback);
    const setPlayback = usePlaybackStore((s) => s.setPlayback);
    const [progress, setProgress] = useState(0);

    const loadPlayback = async () => {
        try {
            const pb = await airstationAPI.getPlayback();
            setPlayback(pb);

            if (!pb.currentTrack) return;
            setProgress((pb.currentTrackElapsed / pb.currentTrack.duration) * 100);
        } catch (error) {
            errNotify(error);
        }
    };

    useEffect(() => {
        const id = setInterval(async () => {
            await loadPlayback();
        }, 1000);

        return () => clearInterval(id);
    }, []);

    return (
        <Paper p="sm" w="100%" h={95} pos="relative">
            <Flex gap="sm">
                <Box w="100%">
                    <Text>{playback?.currentTrack?.name || "Unknown"}</Text>
                    <Space h={10} />
                    <Progress color="air" radius="xl" value={progress} />
                    <Text ta="end" mt={3} c="dimmed">
                        {formatTime(playback?.currentTrackElapsed || 0)}/
                        {formatTime(playback?.currentTrack?.duration || 0)}
                    </Text>
                </Box>
            </Flex>
        </Paper>
    );
};
