import { FC, useEffect, useState } from "react";
import { useDisclosure } from "@mantine/hooks";
import {
    Accordion,
    ActionIcon,
    Button,
    Box,
    Flex,
    Group,
    Modal,
    Textarea,
    TextInput,
    Tooltip,
    Text,
    Slider,
    ColorInput,
    Select,
} from "@mantine/core";
import { IconSettings } from "../icons";
import { useSettingsStore } from "../store/settings";
import { MAX_MOBILE_WIDTH, useIsMobile } from "../hooks/useIsMobile";
import { useForm } from "@mantine/form";
import { StationInfo } from "../api/types";
import { airstationAPI } from "../api";
import { errNotify, okNotify } from "../notifications";

export const SettingsModal: FC<{}> = () => {
    const [opened, { open, close }] = useDisclosure(false);
    const [loading, setLoading] = useState(false);
    const { isMobile } = useIsMobile();

    const interfaceWidth = useSettingsStore((s) => s.interfaceWidth);
    const setInterfaceWidth = useSettingsStore((s) => s.setInterfaceWidth);
    const stationInfo = useForm<StationInfo>({
        initialValues: {
            name: "",
            description: "",
            faviconURL: "",
            logoURL: "",
            location: "",
            timezone: "",
            links: "",
            theme: "",
        },
    });

    const loadStationInfo = async () => {
        setLoading(true);
        try {
            const info = await airstationAPI.getStationInfo();
            stationInfo.setValues(info);
        } catch (error) {
            errNotify(error);
        } finally {
            setLoading(false);
        }
    };

    const saveStationInfo = async (overrides?: Partial<StationInfo>) => {
        setLoading(true);
        try {
            const info = await airstationAPI.editStationInfo({ ...stationInfo.values, ...overrides });
            stationInfo.setValues(info);
            okNotify("Saved successfully");
        } catch (error) {
            errNotify(error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadStationInfo();
    }, []);

    return (
        <>
            <Modal
                title="Settings"
                p={0}
                centered
                size="lg"
                opened={opened}
                onClose={close}
                withCloseButton
                radius="md"
            >
                <Accordion defaultValue="station_info" variant="filled">
                    <Accordion.Item value="station_info">
                        <Accordion.Control>Station info</Accordion.Control>
                        <Accordion.Panel>
                            <TextInput
                                label="Name"
                                placeholder="Enter name"
                                key={stationInfo.key("name")}
                                {...stationInfo.getInputProps("name")}
                            />
                            <Textarea
                                rows={4}
                                maxRows={6}
                                label="Description"
                                placeholder="Enter description"
                                mt="sm"
                                key={stationInfo.key("description")}
                                {...stationInfo.getInputProps("description")}
                            />
                            <Flex mt="sm" gap="xs">
                                <TextInput
                                    w="100%"
                                    label="Location"
                                    placeholder="Enter location"
                                    key={stationInfo.key("location")}
                                    {...stationInfo.getInputProps("location")}
                                />
                                <TextInput
                                    w="100%"
                                    label="Timezone"
                                    placeholder="Enter timezone"
                                    key={stationInfo.key("timezone")}
                                    {...stationInfo.getInputProps("timezone")}
                                />
                            </Flex>
                            <Flex mt="sm" gap="xs">
                                <TextInput
                                    w="100%"
                                    label="Favicon"
                                    placeholder="Enter URL"
                                    key={stationInfo.key("faviconURL")}
                                    {...stationInfo.getInputProps("faviconURL")}
                                />
                                <TextInput
                                    w="100%"
                                    label="Logo"
                                    placeholder="Enter URL"
                                    key={stationInfo.key("logoURL")}
                                    {...stationInfo.getInputProps("logoURL")}
                                />
                            </Flex>

                            <Textarea
                                spellCheck={false}
                                rows={4}
                                maxRows={6}
                                label="Links"
                                placeholder={`Add some links to your socials in format [TITLE](URL)`}
                                mt="sm"
                                key={stationInfo.key("links")}
                                {...stationInfo.getInputProps("links")}
                            />

                            <Group mt="md" justify="flex-end">
                                <Button loading={loading} onClick={() => saveStationInfo()}>
                                    Save
                                </Button>
                            </Group>
                        </Accordion.Panel>
                    </Accordion.Item>

                    <Accordion.Item value="player_theme">
                        <Accordion.Control>Player theme</Accordion.Control>
                        <Accordion.Panel>
                            <PlayerThemeSetup
                                defaultTheme={stationInfo.values.theme}
                                loading={loading}
                                saveStationInfo={saveStationInfo}
                            />
                        </Accordion.Panel>
                    </Accordion.Item>

                    {!isMobile && (
                        <Accordion.Item value="studio_interface">
                            <Accordion.Control>Studio interface</Accordion.Control>
                            <Accordion.Panel>
                                <Box>
                                    <Text size="sm">Interface width in pixels</Text>
                                    <Text size="xs" c="dimmed">
                                        Determines how wide the interface will be.
                                    </Text>
                                    <Slider
                                        value={interfaceWidth}
                                        onChange={setInterfaceWidth}
                                        min={MAX_MOBILE_WIDTH}
                                        max={window.screen.width}
                                        step={10}
                                    />
                                </Box>
                            </Accordion.Panel>
                        </Accordion.Item>
                    )}
                </Accordion>
            </Modal>

            <Tooltip openDelay={500} label="Settings">
                <ActionIcon onClick={open} variant="transparent" size="md">
                    <IconSettings size={18} color="gray" />
                </ActionIcon>
            </Tooltip>
        </>
    );
};

const defaultSwatches = [
    "#2e2e2e",
    "#868e96",
    "#fa5252",
    "#e64980",
    "#be4bdb",
    "#7950f2",
    "#4c6ef5",
    "#228be6",
    "#15aabf",
    "#12b886",
    "#40c057",
    "#82c91e",
    "#fab005",
    "#fd7e14",
];

const customThemeName = "Custom";
const predefinedThemes = {
    Airstation: "#29323c;#485563;#a8a8a8;#ffffff;;",
    "Deep ocean": "#0f2027;#2c5364;#4a90a5;#e6f7ff;#00d4ff;",
    Graphite: "#2d3436;#636e72;#818588;#dfe6e9;;",
    "Sandy beach": "#f7c59f;#efd9b4;#c8a77a;#4a3c2a;#e76f51;",
    "Fresh mint": "#134e5e;#71b280;#5d9170;#e9f5e9;#5af55a;",
    "Misty forest":
        "#2f3f4d;#bcc1cd;#2f3f4d;#ffffff;;https://images.unsplash.com/photo-1487621167305-5d248087c724?q=80\u0026w=1932\u0026auto=format\u0026fit=crop\u0026ixlib=rb-4.1.0\u0026ixid=",
    Hackerman: "#000000;#000000;#04e600;#04e600;#04e600;",
    "Just dark": "#000000;#000000;#a8a8a8;#ffffff;;",
    "Just light": "#ffffff;#ffffff;#a8a8a8;#000000;;",
};

const PlayerThemeSetup: FC<{
    defaultTheme: string;
    loading: boolean;
    saveStationInfo: (overrides?: Partial<StationInfo> | undefined) => Promise<void>;
}> = ({ defaultTheme, loading, saveStationInfo }) => {
    const parsedDefaultTheme = defaultTheme.split(";");

    const [theme, setTheme] = useState<string>(customThemeName);
    const [bgStart, setBgStart] = useState(parsedDefaultTheme[0] || "");
    const [bgEnd, setBgEnd] = useState(parsedDefaultTheme[1] || "");
    const [iconsColor, setIconsColor] = useState(parsedDefaultTheme[2] || "");
    const [textColor, setTextColor] = useState(parsedDefaultTheme[3] || "");
    const [accentColor, setAccentColor] = useState(parsedDefaultTheme[4] || "");
    const [bgImage, setBgImage] = useState(parsedDefaultTheme[5] || "");

    const getThemeString = () => {
        return `${bgStart};${bgEnd};${iconsColor};${textColor};${accentColor};${bgImage}`;
    };

    const handleSave = async () => {
        await saveStationInfo({ theme: getThemeString() });
    };

    const handlePredefinedTheme = (value: string | null) => {
        if (value == null) {
            setTheme(customThemeName);
            return;
        }

        const pt = predefinedThemes[value as keyof typeof predefinedThemes];
        if (pt) {
            const parsed = pt.split(";");
            setBgStart(parsed[0]);
            setBgEnd(parsed[1]);
            setIconsColor(parsed[2]);
            setTextColor(parsed[3]);
            setAccentColor(parsed[4]);
            setBgImage(parsed[5]);
        }

        setTheme(value);
    };

    const defineThemeName = () => {
        const themeString = `${bgStart};${bgEnd};${iconsColor};${textColor};${accentColor};${bgImage}`;

        if (Object.values(predefinedThemes).includes(themeString)) {
            setTheme(
                Object.keys(predefinedThemes).find(
                    (key) => predefinedThemes[key as keyof typeof predefinedThemes] === themeString,
                ) || customThemeName,
            );
        } else {
            setTheme(customThemeName);
        }
    };

    useEffect(() => {
        defineThemeName();
    }, [bgStart, bgEnd, iconsColor, textColor, accentColor, bgImage]);

    return (
        <Flex direction="column" gap="sm">
            <Text size="sm" c="dimmed">
                Here you can customize the appearance of the station's player page.
            </Text>

            <Select
                label="Select pre-defined theme"
                placeholder="Pre-defined theme"
                value={theme}
                onChange={handlePredefinedTheme}
                data={[...Object.keys(predefinedThemes), customThemeName]}
            />

            <Text size="xs" c="dimmed">
                Below you can define the color for the background. Two colors are needed to create a gradient effect. If
                you need a solid color, just repeat it 2 times.
            </Text>
            <Flex gap="sm">
                <ColorInput
                    w="100%"
                    label="Background start"
                    placeholder="Bottom color"
                    format="hex"
                    value={bgStart}
                    onChange={setBgStart}
                    swatches={defaultSwatches}
                />
                <ColorInput
                    w="100%"
                    label="Background end"
                    placeholder="Top color"
                    format="hex"
                    value={bgEnd}
                    onChange={setBgEnd}
                    swatches={defaultSwatches}
                />
            </Flex>
            <Flex gap="sm">
                <ColorInput
                    w="100%"
                    label="Icons color"
                    placeholder="Color for icons"
                    format="hex"
                    value={iconsColor}
                    onChange={setIconsColor}
                    swatches={defaultSwatches}
                />
                <ColorInput
                    w="100%"
                    label="Text color"
                    placeholder="Color for text"
                    format="hex"
                    value={textColor}
                    onChange={setTextColor}
                    swatches={defaultSwatches}
                />
            </Flex>

            <ColorInput
                w="100%"
                label="Accent color"
                placeholder="Color for music visualizer"
                description="The color for the music visualizer when the play button is pressed. If the value is empty, the visualizer will sparkle with multicolored shades."
                format="hex"
                value={accentColor}
                onChange={setAccentColor}
                swatches={defaultSwatches}
            />
            <TextInput
                w="100%"
                label="Background image URL"
                description="A link to the image that will be used instead of the colored background."
                placeholder="Enter URL"
                value={bgImage}
                onChange={(e) => setBgImage(e.target.value)}
            />
            <Group mt="sm" justify="flex-end">
                <Button loading={loading} onClick={handleSave}>
                    Save
                </Button>
            </Group>
        </Flex>
    );
};
