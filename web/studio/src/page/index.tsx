import { lazy, Suspense } from "react";
import { useIsMobile } from "../hooks/useIsMobile";

const DesktopPage = lazy(() => import("./DesktopPage"));
const MobilePage = lazy(() => import("./MobilePage"));

export const Page = () => {
    const { isMobile, width } = useIsMobile();
    const PageComponent = isMobile ? MobilePage : DesktopPage;

    return (
        <Suspense fallback={null}>
            <PageComponent windowWidth={width} />
        </Suspense>
    );
};
