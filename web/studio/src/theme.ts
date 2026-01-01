import { createTheme, LoadingOverlay, Overlay } from "@mantine/core";

export const theme = createTheme({
    fontFamily: '"Exo 2", serif',

    colors: {
        dark: [
            "#f3f5f7",
            "#e7e7e7",
            "#cacccf",
            "#aab0b7",
            "#4d555e",
            "#7e8997",
            "#262f3a",
            "#2e3944",
            "#566373",
            "#29323c",
        ],
        air: [
            "#dffbff",
            "#caf2ff",
            "#99e2ff",
            "#64d2ff",
            "#3cc4fe",
            "#23bcfe",
            "#09b8ff",
            "#00a1e4",
            "#008fcd",
            "#007cb6",
        ],
    },
    primaryColor: "air",
    defaultGradient: { from: "#29323c", to: "#485563", deg: -180 },
    components: {
        LoadingOverlay: LoadingOverlay.extend({
            defaultProps: {
                overlayProps: {},
            },
        }),
        Overlay: Overlay.extend({
            defaultProps: {
                bg: "rgba(0, 0, 0, 0.4)",
            },
        }),
    },
});
