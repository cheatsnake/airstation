import { Container, SimpleGrid } from "@mantine/core";
import { Playback } from "./Playback";
import { TrackQueue } from "./TracksQueue";
import { TrackLibrary } from "./TracksLibrary";

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
