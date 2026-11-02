import { Container, Flex, Tabs } from "@mantine/core";
import { FC, useState } from "react";
import { Playback } from "./Playback";
import { TrackLibrary } from "./TracksLibrary";
import { TrackQueue } from "./TracksQueue";
import { Console } from "./Console";
import { Schedule } from "./Schedule";
import { useSettingsStore } from "../store/settings";

const DesktopPage: FC<{ windowWidth: number }> = ({ windowWidth }) => {
    const interfaceWidth = useSettingsStore((s) => s.interfaceWidth);
    const [tab, setTab] = useState<string | null>("console");

    const defineWidth = () => {
        if (interfaceWidth) return interfaceWidth;
        return windowWidth >= 2400 ? "xl" : "lg";
    };

    return (
        <Container size={defineWidth()}>
            <Flex p="sm" direction="column" justify="center" align="center" h="100vh" gap="sm">
                <Playback />

                <Tabs
                    value={tab}
                    onChange={setTab}
                    variant="pills"
                    w="100%"
                    style={{ flex: 1, minHeight: 0, display: "flex", flexDirection: "column" }}
                >
                    <Tabs.List>
                        <Tabs.Tab value="console">Console</Tabs.Tab>
                        <Tabs.Tab value="queue">Queue</Tabs.Tab>
                        <Tabs.Tab value="library">Library</Tabs.Tab>
                        <Tabs.Tab value="schedule">Schedule</Tabs.Tab>
                    </Tabs.List>

                    <Tabs.Panel value="console" pt="sm" style={{ flex: 1, minHeight: 0 }}>
                        <Console />
                    </Tabs.Panel>
                    <Tabs.Panel value="queue" pt="sm" style={{ flex: 1, minHeight: 0 }}>
                        <TrackQueue />
                    </Tabs.Panel>
                    <Tabs.Panel value="library" pt="sm" style={{ flex: 1, minHeight: 0 }}>
                        <TrackLibrary />
                    </Tabs.Panel>
                    <Tabs.Panel value="schedule" pt="sm" style={{ flex: 1, minHeight: 0 }}>
                        <Schedule />
                    </Tabs.Panel>
                </Tabs>
            </Flex>
        </Container>
    );
};

export default DesktopPage;
