import { Container, Flex, SimpleGrid } from "@mantine/core";
import { Playback } from "./Playback";
import { TrackQueue } from "./TracksQueue";
import { TrackLibrary } from "./TracksLibrary";

export const Page = () => {
    return (
        <Container size="lg">
            <Flex p="sm" direction="column" justify="center" align="center" h="100vh">
                <Playback />

                <SimpleGrid cols={{ base: 1, sm: 2 }} mt="sm" w="100%">
                    <TrackQueue />
                    <TrackLibrary />
                </SimpleGrid>
            </Flex>
        </Container>
    );
};
