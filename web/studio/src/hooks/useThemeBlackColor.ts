import { useMantineColorScheme } from "@mantine/core";

export function useThemeBlackColor() {
    const { colorScheme } = useMantineColorScheme();
    return colorScheme === "dark" ? "gray" : "black";
}
