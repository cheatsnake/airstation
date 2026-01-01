import { useViewportSize } from "@mantine/hooks";

export const MAX_MOBILE_WIDTH = 800;

export const useIsMobile = () => {
    const { width } = useViewportSize();
    const isMobile = width <= MAX_MOBILE_WIDTH;
    return { width, isMobile };
};
