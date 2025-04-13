import { Button, Flex } from "@mantine/core";
import { FC } from "react";

interface MobileBarProps {
    activeBar: string;
    setActiveBar: React.Dispatch<React.SetStateAction<string>>;
}

export const MOBILE_BARS = ["Playback", "Queue", "Tracks", "Settings"];

export const MobileBar: FC<MobileBarProps> = ({ activeBar, setActiveBar }) => {
    return (
        <Flex w="100%" justify="space-around" align="center" px="sm">
            {MOBILE_BARS.map((bar) => (
                <Button
                    my="sm"
                    variant="transparent"
                    c={bar === activeBar ? "air" : "white"}
                    onClick={() => setActiveBar(bar)}
                >
                    {bar}
                </Button>
            ))}
        </Flex>
    );
};
