import { Container, Flex, SimpleGrid } from "@mantine/core";
import { FC } from "react";
import { Playback } from "./Playback";
import { TrackLibrary } from "./TracksLibrary";
import { TrackQueue } from "./TracksQueue";

const DesktopPage: FC<{ windowWidth: number }> = ({ windowWidth }) => {
    return (
        <Container size={windowWidth > 2500 ? "xl" : "lg"}>
            <Flex p="sm" direction="column" justify="center" align="center" h="100vh">
                <Playback />

                <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="sm" mt="sm" w="100%">
                    <TrackQueue />
                    <TrackLibrary />
                </SimpleGrid>
            </Flex>
        </Container>
    );
};

export default DesktopPage;
