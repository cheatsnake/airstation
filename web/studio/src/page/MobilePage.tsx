import { Flex } from "@mantine/core";
import { useState } from "react";
import { MobileBar } from "./MobileBar";
import { Playback } from "./Playback";
import { TrackLibrary } from "./TracksLibrary";
import { TrackQueue } from "./TracksQueue";

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

export default MobilePage;
