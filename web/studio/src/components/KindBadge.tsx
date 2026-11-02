import { Badge } from "@mantine/core";
import { FC } from "react";
import { trackKindColor, trackKindLabel } from "../api/types";

export const KindBadge: FC<{ kind?: string; size?: "xs" | "sm" | "md" }> = ({ kind, size = "xs" }) => {
    if (!kind || kind === "music") return null;

    return (
        <Badge size={size} variant="light" color={trackKindColor(kind)} style={{ flexShrink: 0 }}>
            {trackKindLabel(kind)}
        </Badge>
    );
};
