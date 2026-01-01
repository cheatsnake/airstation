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
} from "@mantine/core";
import { IconSettings } from "../icons";
import { useSettingsStore } from "../store/settings";
import { MAX_MOBILE_WIDTH, useIsMobile } from "../hooks/useIsMobile";
import { useForm } from "@mantine/form";
import { StationInfo } from "../api/types";
import { airstationAPI } from "../api";
import { errNotify } from "../notifications";

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

    const saveStationInfo = async () => {
        setLoading(true);
        try {
            const info = await airstationAPI.editStationInfo(stationInfo.values);
            stationInfo.setValues(info);
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
                                <Button loading={loading} onClick={saveStationInfo}>
                                    Save
                                </Button>
                            </Group>
                        </Accordion.Panel>
                    </Accordion.Item>

                    {!isMobile && (
                        <Accordion.Item value="studio_interface">
                            <Accordion.Control>Studio interface</Accordion.Control>
                            <Accordion.Panel>
                                <Box>
                                    <Text size="sm">Message width in pixels</Text>
                                    <Text size="xs" c="dimmed">
                                        Determines how wide the chat area with messages will be.
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
