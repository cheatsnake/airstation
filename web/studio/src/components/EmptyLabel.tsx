import { Flex, Text } from "@mantine/core";
import { FC } from "react";

export const EmptyLabel: FC<{ label: string }> = ({ label }) => {
    return (
        <Flex justify="center" align="center" w="100%" h="100%">
            <Text fz="lg" c="dimmed">
                {label}
            </Text>
        </Flex>
    );
};
