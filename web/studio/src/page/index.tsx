import { ActionIcon, Box, Container, Flex, Paper, Progress, SimpleGrid, Space, Text, TextInput } from "@mantine/core";
import { FC, useEffect, useState } from "react";
import { airstationAPI } from "../api";
import { usePlaybackStore } from "../store/playback";
import { useTrackQueueStore } from "../store/track-queue";
import { useTracksStore } from "../store/tracks";
import { IconPlayerPlayFilled } from "../icons/IconPlayerPlayFilled";
import { IconPlayerStopFilled } from "../icons/IconPlayerStopFilled";
import { formatTime } from "../utils/time";

export const Page = () => {
  return (
    <Container p="sm">
      <Playback />

      <SimpleGrid cols={{ base: 1, sm: 2 }} mt="sm">
        <TrackQueue />
        <TrackLibrary />
      </SimpleGrid>
    </Container>
  );
};

const Playback: FC<{}> = () => {
  const playback = usePlaybackStore((s) => s.playback);
  const setPlayback = usePlaybackStore((s) => s.setPlayback);
  const [progress, setProgress] = useState(0);

  const loadPlayback = async () => {
    try {
      const pb = await airstationAPI.getPlayback();
      setPlayback(pb);
      setProgress((pb.currentTrackElapsed / pb.currentTrack.duration) * 100);
    } catch (error) {
      console.log(error);
    }
  };

  useEffect(() => {
    const id = setInterval(async () => {
      await loadPlayback();
    }, 1000);

    return () => clearInterval(id);
  }, []);

  return (
    <Paper p="sm">
      <Flex gap="sm">
        <ActionIcon variant="subtle" color="white" size="xl" aria-label="Settings">
          {playback?.IsPlaying ? (
            <IconPlayerPlayFilled style={{ width: "70%", height: "70%" }} fill="white" stroke={1.5} />
          ) : (
            <IconPlayerStopFilled style={{ width: "70%", height: "70%" }} fill="white" stroke={1.5} />
          )}
        </ActionIcon>
        <Box w="100%">
          <Text>{playback?.currentTrack.name}</Text>
          <Space h={10} />
          <Progress color="air" radius="xl" value={progress} />
          <Text ta="end" mt={3} c="dimmed">
            {formatTime(playback?.currentTrackElapsed || 0)}/{formatTime(playback?.currentTrack.duration || 0)}
          </Text>
        </Box>
      </Flex>
    </Paper>
  );
};

const TrackQueue: FC<{}> = () => {
  const playback = usePlaybackStore((s) => s.playback);
  const queue = useTrackQueueStore((s) => s.queue);
  const setQueue = useTrackQueueStore((s) => s.setQueue);

  const loadQueue = async () => {
    try {
      const q = await airstationAPI.getQueue();
      setQueue(q);
    } catch (error) {
      console.log(error);
    }
  };

  useEffect(() => {
    loadQueue();
  }, []);

  return (
    <Paper withBorder mih={300} p="sm">
      <Flex justify="space-between" align="center">
        <Text fw={700} size="lg">
          Live queue
        </Text>
        <Text c="dimmed">{`${queue.length - 1} ${queue.length > 2 ? "tracks" : "track"}`}</Text>
      </Flex>
      <Space h={12} />

      <Flex direction="column" gap="sm">
        {queue.length > 1 ? (
          queue
            .filter((t) => playback?.currentTrack.id != t.id)
            .map((track) => (
              <Paper p="xs" withBorder key={track.id}>
                {track.name}
              </Paper>
            ))
        ) : (
          <EmptyLabel label={"Queue is empty"} />
        )}
      </Flex>
    </Paper>
  );
};

const TrackLibrary: FC<{}> = () => {
  const tracks = useTracksStore((s) => s.tracks);
  const queue = useTrackQueueStore((s) => s.queue);
  const setTracks = useTracksStore((s) => s.setTracks);

  const loadTracks = async (page = 1, limit = 20, search = "") => {
    try {
      const page = await airstationAPI.getTracks(1, 20, "");
      setTracks(page.tracks);
    } catch (error) {
      console.log(error);
    }
  };

  const isTrackInQueue = (trackID: string) => {
    return queue.map((t) => t.id).includes(trackID);
  };

  useEffect(() => {
    loadTracks();
  }, []);

  return (
    <Paper withBorder mih={100} p="xs">
      <Flex justify="space-between" align="center">
        <Text fw={700} size="lg">
          Tracks library
        </Text>
        <Text c="dimmed">{`${tracks.length} ${tracks.length > 1 ? "tracks" : "track"}`}</Text>
      </Flex>
      <Space h={12} />
      <TextInput placeholder="Search" />
      <Space h={16} />
      <Flex direction="column" gap="sm">
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

const EmptyLabel: FC<{ label: string }> = ({ label }) => {
  return (
    <Flex justify="center" align="center" w="100%" h="100%">
      <Text fz="lg" c="dimmed">
        {label}
      </Text>
    </Flex>
  );
};
