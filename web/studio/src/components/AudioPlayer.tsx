import { ActionIcon, Box, Flex, Group, Paper, Progress, Text, Tooltip } from "@mantine/core";
import React, { useEffect, useRef } from "react";
import { IconPlayerPlayFilled } from "../icons/IconPlayerPlayFilled";
import { IconPlayerStopFilled } from "../icons/IconPlayerStopFilled";
import { Track } from "../api/types";
import { API_HOST } from "../api";
import { formatTime } from "../utils/time";
import { useThrottledState } from "@mantine/hooks";

interface AudioPlayerProps {
    track: Track;
    isPlaying: boolean;
    togglePlaying: () => void;
}

export const AudioPlayer: React.FC<AudioPlayerProps> = ({ track, isPlaying, togglePlaying }) => {
    const audioRef = useRef<HTMLAudioElement>(null);
    const [progress, setProgress] = useThrottledState(0, 500);
    const [cursorPos, setCursorPos] = useThrottledState(0, 300);

    const handleProgressClick = (e: React.MouseEvent<HTMLDivElement>) => {
        if (audioRef.current) {
            const rect = e.currentTarget.getBoundingClientRect();
            const clickPosition = (e.clientX - rect.left) / rect.width;
            const newTime = clickPosition * audioRef.current.duration;
            audioRef.current.currentTime = newTime;
            setProgress(clickPosition * 100);
            if (!isPlaying) togglePlaying();
        }
    };

    const handleTimeUpdate = () => {
        if (audioRef.current) {
            const currentTime = audioRef.current.currentTime;
            const duration = audioRef.current.duration;
            setProgress((currentTime / duration) * 100);
        }
    };

    const handleAudioEnd = () => {
        if (audioRef.current) {
            audioRef.current.pause();
            audioRef.current.currentTime = 0;
            togglePlaying();
            setProgress(0);
        }
    };

    useEffect(() => {
        if (audioRef.current) {
            if (isPlaying) {
                audioRef.current.play();
            } else {
                audioRef.current.pause();
            }
        }
    }, [isPlaying]);

    return (
        <Paper>
            <audio
                ref={audioRef}
                src={`${API_HOST}/${track.path}`}
                preload="none"
                onTimeUpdate={handleTimeUpdate}
                onEnded={handleAudioEnd}
            />

            <Text>{track.name}</Text>
            <Flex gap="sm" align="center">
                <ActionIcon onClick={togglePlaying} variant="subtle" color="white" size="sm" aria-label="Settings">
                    {isPlaying ? (
                        <IconPlayerStopFilled fill="white" stroke={1.5} />
                    ) : (
                        <IconPlayerPlayFilled fill="white" stroke={1.5} />
                    )}
                </ActionIcon>

                <Box w="100%" mt="xs" style={{ cursor: "pointer" }}>
                    <Tooltip.Floating label={formatTime(track.duration * cursorPos)}>
                        <Progress
                            onMouseMove={(e) => {
                                const rect = e.currentTarget.getBoundingClientRect();
                                setCursorPos(Math.abs((e.clientX - rect.left) / rect.width));
                            }}
                            onClick={handleProgressClick}
                            value={progress}
                        />
                    </Tooltip.Floating>

                    <Group align="end">
                        <Text ta="end" mt={3} c="dimmed" size="sm">
                            {formatTime((progress / 100) * track.duration || 0)}/{formatTime(track.duration || 0)}
                        </Text>
                    </Group>
                </Box>
            </Flex>
        </Paper>
    );
};
