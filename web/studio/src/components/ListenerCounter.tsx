import { FC, useEffect, useState } from "react";
import { Flex, Text, Tooltip } from "@mantine/core";
import { IconHeadphones } from "../icons";
import { EVENTS, useEventSourceStore } from "../store/events";

export const ListenerCounter: FC<{ size?: number; large?: boolean }> = ({ size = 18, large }) => {
    const [count, setCount] = useState(0);
    const addEventHandler = useEventSourceStore((s) => s.addEventHandler);

    useEffect(() => {
        addEventHandler(EVENTS.countListeners, (msg: MessageEvent<string>) => {
            setCount(Number(msg.data));
        });
    }, []);

    return (
        <Tooltip openDelay={500} label="Listeners connected right now">
            <Flex gap={5} justify="center" align="center" opacity={large ? 1 : 0.5}>
                <IconHeadphones size={large ? 26 : size} />
                <Text size={large ? "xl" : undefined} fw={large ? 700 : undefined}>
                    {count}
                </Text>
            </Flex>
        </Tooltip>
    );
};
