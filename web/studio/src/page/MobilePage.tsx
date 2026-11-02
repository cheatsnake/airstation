import { Flex } from "@mantine/core";
import { useState } from "react";
import { MobileBar } from "./MobileBar";
import { Playback } from "./Playback";
import { TrackLibrary } from "./TracksLibrary";
import { TrackQueue } from "./TracksQueue";
import { Console } from "./Console";
import { Schedule } from "./Schedule";

const MobilePage = () => {
    const [activeBar, setActiveBar] = useState("Console");
    const isVisible = (bar: string) => (bar === activeBar ? "block" : "none");

    return (
        <Flex direction="column" h="100vh">
            <div style={{ flex: 1, display: isVisible("Console"), overflowY: "auto" }}>
                <Console isMobile />
            </div>
            <div style={{ flex: 1, display: isVisible("Playback") }}>
                <Playback isMobile />
            </div>
            <div style={{ flex: 1, display: isVisible("Queue") }}>
                <TrackQueue isMobile />
            </div>
            <div style={{ flex: 1, display: isVisible("Tracks") }}>
                <TrackLibrary isMobile />
            </div>
            <div style={{ flex: 1, display: isVisible("Schedule") }}>
                <Schedule isMobile />
            </div>

            <MobileBar activeBar={activeBar} setActiveBar={setActiveBar} />
        </Flex>
    );
};

export default MobilePage;
