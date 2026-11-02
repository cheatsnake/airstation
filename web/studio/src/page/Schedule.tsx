import { FC, useEffect, useState } from "react";
import {
    ActionIcon,
    Box,
    Button,
    Flex,
    Group,
    Modal,
    NumberInput,
    Paper,
    Select,
    SimpleGrid,
    Stack,
    Switch,
    Text,
    TextInput,
    Title,
    Tooltip,
} from "@mantine/core";
import { modals } from "@mantine/modals";
import { useDisclosure } from "@mantine/hooks";
import { useScheduleStore } from "../store/schedule";
import { KindBadge } from "../components/KindBadge";
import { ScheduleRule, TrackKind } from "../api/types";
import { EmptyLabel } from "../components/EmptyLabel";
import { errNotify, okNotify, warnNotify } from "../notifications";
import { IconReload, IconTrash } from "../icons";
import styles from "./styles.module.css";

const KIND_OPTIONS = [
    { value: "jingle", label: "Jingle" },
    { value: "ad", label: "Ad" },
    { value: "station_id", label: "Station ID" },
];

interface RuleForm {
    name: string;
    kind: TrackKind;
    everyNTracks: number | string;
    clockEnabled: boolean;
    atMinute: number | string;
    enabled: boolean;
}

const emptyForm: RuleForm = {
    name: "",
    kind: "jingle",
    everyNTracks: 3,
    clockEnabled: false,
    atMinute: 0,
    enabled: true,
};

export const Schedule: FC<{ isMobile?: boolean }> = ({ isMobile }) => {
    const rules = useScheduleStore((s) => s.rules);
    const fetchRules = useScheduleStore((s) => s.fetchRules);
    const createRule = useScheduleStore((s) => s.createRule);
    const editRule = useScheduleStore((s) => s.editRule);
    const deleteRule = useScheduleStore((s) => s.deleteRule);

    const [opened, { open, close }] = useDisclosure(false);
    const [editing, setEditing] = useState<ScheduleRule | null>(null);
    const [form, setForm] = useState<RuleForm>(emptyForm);
    const [busy, setBusy] = useState(false);

    useEffect(() => {
        fetchRules().catch((error) => errNotify(error));
    }, []);

    const openCreate = () => {
        setEditing(null);
        setForm(emptyForm);
        open();
    };

    const openEdit = (rule: ScheduleRule) => {
        setEditing(rule);
        setForm({
            name: rule.name,
            kind: rule.kind,
            everyNTracks: rule.everyNTracks || 0,
            clockEnabled: rule.atMinute >= 0,
            atMinute: rule.atMinute >= 0 ? rule.atMinute : 0,
            enabled: rule.enabled,
        });
        open();
    };

    const submit = async () => {
        const everyNTracks = Number(form.everyNTracks) || 0;
        const atMinute = form.clockEnabled ? Number(form.atMinute) : -1;

        if (form.name.trim().length < 2) {
            warnNotify("Give the rule a name (at least 2 characters)");
            return;
        }
        if (everyNTracks <= 0 && atMinute < 0) {
            warnNotify("Enable a track rotation and/or a clock trigger");
            return;
        }

        setBusy(true);
        try {
            const draft = {
                name: form.name.trim(),
                kind: form.kind,
                everyNTracks,
                atMinute,
                enabled: form.enabled,
            };
            if (editing) {
                await editRule(editing.id, draft);
                okNotify("Schedule rule updated");
            } else {
                await createRule(draft);
                okNotify("Schedule rule created");
            }
            close();
        } catch (error) {
            errNotify(error);
        } finally {
            setBusy(false);
        }
    };

    const handleToggle = async (rule: ScheduleRule, enabled: boolean) => {
        try {
            await editRule(rule.id, { ...rule, enabled });
        } catch (error) {
            errNotify(error);
        }
    };

    const confirmDelete = (rule: ScheduleRule) => {
        modals.openConfirmModal({
            title: "Delete schedule rule",
            centered: true,
            cancelProps: { variant: "light", color: "gray" },
            children: <Text size="sm">{`Delete “${rule.name}”? Interludes will stop being inserted automatically.`}</Text>,
            labels: { confirm: "Delete", cancel: "Cancel" },
            confirmProps: { color: "red" },
            onConfirm: async () => {
                try {
                    await deleteRule(rule.id);
                    okNotify("Schedule rule deleted");
                } catch (error) {
                    errNotify(error);
                }
            },
        });
    };

    return (
        <Paper radius="md" className={styles.transparent_paper} h="100%">
            <Flex p="sm" direction="column" h={isMobile ? "calc(100vh - 70px)" : "100%"}>
                <Group justify="space-between" align="center">
                    <Group gap="xs">
                        <Title order={3}>Schedule</Title>
                        <Text c="dimmed">{`${rules.filter((r) => r.enabled).length}/${rules.length} active`}</Text>
                    </Group>
                    <Group gap="xs">
                        <Tooltip label="Reload schedule">
                            <ActionIcon variant="transparent" onClick={() => void fetchRules()}>
                                <IconReload size={18} color="gray" />
                            </ActionIcon>
                        </Tooltip>
                        <Button variant="light" onClick={openCreate}>
                            New rule
                        </Button>
                    </Group>
                </Group>

                <Text size="sm" c="dimmed" mt={4}>
                    Rules automatically insert jingles, ads and station IDs into the program log. Clock rules fire once
                    per hour; rotation rules fire after a number of music tracks.
                </Text>

                <Box mt="md" style={{ overflowY: "auto", flex: 1 }}>
                    <SimpleGrid cols={{ base: 1, md: 2 }} spacing="sm">
                        {rules.map((rule) => (
                            <Paper key={rule.id} p="sm" radius="md" withBorder bg="rgba(0,0,0,0.18)">
                                <Group justify="space-between" align="flex-start" wrap="nowrap">
                                    <Group gap="xs" wrap="nowrap">
                                        <KindBadge kind={rule.kind} size="sm" />
                                        <Text fw={600} lineClamp={1}>
                                            {rule.name}
                                        </Text>
                                    </Group>
                                    <Switch
                                        checked={rule.enabled}
                                        onChange={(e) => void handleToggle(rule, e.currentTarget.checked)}
                                        size="sm"
                                    />
                                </Group>
                                <Text size="sm" c="dimmed" mt={4}>
                                    {describeRule(rule.everyNTracks, rule.atMinute)}
                                </Text>
                                <Group justify="flex-end" gap="xs" mt="xs">
                                    <Button size="compact-sm" variant="subtle" onClick={() => openEdit(rule)}>
                                        Edit
                                    </Button>
                                    <Button
                                        size="compact-sm"
                                        variant="subtle"
                                        color="red"
                                        leftSection={<IconTrash size={14} />}
                                        onClick={() => confirmDelete(rule)}
                                    >
                                        Delete
                                    </Button>
                                </Group>
                            </Paper>
                        ))}
                    </SimpleGrid>
                    {!rules.length ? <EmptyLabel label="No schedule rules yet" /> : null}
                </Box>
            </Flex>

            <Modal opened={opened} onClose={close} title={editing ? "Edit schedule rule" : "New schedule rule"} centered>
                <Stack gap="sm">
                    <TextInput
                        label="Name"
                        placeholder="e.g. Top of the hour ID"
                        value={form.name}
                        onChange={(e) => setForm({ ...form, name: e.currentTarget.value })}
                    />
                    <Select
                        label="Asset kind"
                        data={KIND_OPTIONS}
                        value={form.kind}
                        onChange={(value) => setForm({ ...form, kind: (value as TrackKind) ?? "jingle" })}
                        allowDeselect={false}
                    />
                    <NumberInput
                        label="Every N music tracks"
                        description="0 disables the rotation trigger"
                        min={0}
                        max={1000}
                        value={form.everyNTracks}
                        onChange={(value) => setForm({ ...form, everyNTracks: value })}
                    />
                    <Switch
                        label="Clock trigger"
                        description="Insert the asset at a specific minute of the hour"
                        checked={form.clockEnabled}
                        onChange={(e) => setForm({ ...form, clockEnabled: e.currentTarget.checked })}
                    />
                    {form.clockEnabled ? (
                        <NumberInput
                            label="Minute of the hour"
                            min={0}
                            max={59}
                            value={form.atMinute}
                            onChange={(value) => setForm({ ...form, atMinute: value })}
                        />
                    ) : null}
                    <Switch
                        label="Enabled"
                        checked={form.enabled}
                        onChange={(e) => setForm({ ...form, enabled: e.currentTarget.checked })}
                    />
                    <Group justify="flex-end">
                        <Button variant="light" color="gray" onClick={close}>
                            Cancel
                        </Button>
                        <Button loading={busy} onClick={submit}>
                            {editing ? "Save" : "Create"}
                        </Button>
                    </Group>
                </Stack>
            </Modal>
        </Paper>
    );
};

const describeRule = (everyNTracks: number, atMinute: number) => {
    const parts: string[] = [];
    if (everyNTracks > 0) parts.push(`every ${everyNTracks} music track${everyNTracks > 1 ? "s" : ""}`);
    if (atMinute >= 0) parts.push(`at :${String(atMinute).padStart(2, "0")} each hour`);
    return parts.join(" • ") || "no trigger";
};
