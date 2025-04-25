import { useViewportSize } from "@mantine/hooks";
import { lazy, Suspense } from "react";

const DesktopPage = lazy(() => import("./DesktopPage"));
const MobilePage = lazy(() => import("./MobilePage"));
const MAX_MOBILE_WIDTH = 800;

export const Page = () => {
    const { width: windowWidth } = useViewportSize();
    const PageComponent = windowWidth > MAX_MOBILE_WIDTH ? DesktopPage : MobilePage;

    return (
        <Suspense fallback={null}>
            <PageComponent windowWidth={windowWidth} />
        </Suspense>
    );
};
