import { FC } from "react";
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

export const SettingsModal: FC<{}> = () => {
    const [opened, { open, close }] = useDisclosure(false);
    const { isMobile } = useIsMobile();
    const interfaceWidth = useSettingsStore((s) => s.interfaceWidth);
    const setInterfaceWidth = useSettingsStore((s) => s.setInterfaceWidth);

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
                            <TextInput label="Name" placeholder="Enter name" />
                            <Textarea
                                rows={4}
                                maxRows={6}
                                label="Description"
                                placeholder="Enter description"
                                mt="sm"
                            />
                            <Flex mt="sm" gap="xs">
                                <TextInput w="100%" label="Location" placeholder="Enter location" />
                                <TextInput w="100%" label="Timezone" placeholder="Enter timezone" />
                            </Flex>
                            <Flex mt="sm" gap="xs">
                                <TextInput w="100%" label="Favicon" placeholder="Enter URL" />
                                <TextInput w="100%" label="Logo" placeholder="Enter URL" />
                            </Flex>

                            <Textarea
                                rows={4}
                                maxRows={6}
                                label="Links"
                                placeholder={`Add some links to your socials in format [TITLE](URL)`}
                                mt="sm"
                            />

                            <Group mt="md" justify="flex-end">
                                <Button>Save</Button>
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
