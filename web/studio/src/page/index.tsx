import { Container, Flex, SimpleGrid } from "@mantine/core";
import { Playback } from "./Playback";
import { TrackQueue } from "./TracksQueue";
import { TrackLibrary } from "./TracksLibrary";
import { useViewportSize } from "@mantine/hooks";
import { MobileBar } from "./MobileBar";
import { FC, useState } from "react";

const MAX_MOBILE_WIDTH = 800;

export const Page = () => {
    const { width: windowWidth } = useViewportSize();

    return <>{windowWidth > MAX_MOBILE_WIDTH ? <DesktopPage windowWidth={windowWidth} /> : <MobilePage />}</>;
};

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

const MobilePage = () => {
    const [activeBar, setActiveBar] = useState("Playback");
    const isVisible = (bar: string) => (bar === activeBar ? "block" : "none");

    return (
        <Flex direction="column" h="100vh">
            <div style={{ flex: 1, display: isVisible("Playback") }}>
                <Playback isMobile />
            </div>
            <div style={{ flex: 1, display: isVisible("Queue") }}>
                <TrackQueue isMobile />
            </div>
            <div style={{ flex: 1, display: isVisible("Tracks") }}>
                <TrackLibrary isMobile />
            </div>
            <div style={{ flex: 1, display: isVisible("Settings") }}>
                <></>
            </div>

            <MobileBar activeBar={activeBar} setActiveBar={setActiveBar} />
        </Flex>
    );
};
