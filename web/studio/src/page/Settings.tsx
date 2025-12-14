import { Accordion, ActionIcon, Button, Flex, Group, Modal, Textarea, TextInput, Tooltip } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { FC } from "react";
import { IconSettings } from "../icons";

export const SettingsModal: FC<{}> = () => {
    const [opened, { open, close }] = useDisclosure(false);

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

                            <Group mt="md" justify="flex-end">
                                <Button>Save</Button>
                            </Group>
                        </Accordion.Panel>
                    </Accordion.Item>
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
