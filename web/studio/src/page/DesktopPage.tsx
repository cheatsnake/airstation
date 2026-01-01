import { Container, Flex, SimpleGrid } from "@mantine/core";
import { FC } from "react";
import { Playback } from "./Playback";
import { TrackLibrary } from "./TracksLibrary";
import { TrackQueue } from "./TracksQueue";
import { useSettingsStore } from "../store/settings";

const DesktopPage: FC<{ windowWidth: number }> = ({ windowWidth }) => {
    const interfaceWidth = useSettingsStore((s) => s.interfaceWidth);
    const defineWidth = () => {
        if (interfaceWidth) return interfaceWidth;
        return windowWidth >= 2400 ? "xl" : "lg";
    };

    return (
        <Container size={defineWidth()}>
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
