import { createTheme, MantineProvider } from "@mantine/core";
import { Page } from "./page";
import { Notifications } from "@mantine/notifications";
import { ModalsProvider } from "@mantine/modals";
import { AuthGuard } from "./page/AuthGuard";

const theme = createTheme({
    fontFamily: '"Exo 2", serif',
    colors: {
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
});

const App = () => {
    return (
        <MantineProvider defaultColorScheme="dark" theme={theme}>
            <ModalsProvider>
                <Notifications position="bottom-right" autoClose={7000} />
                <AuthGuard>
                    <Page />
                </AuthGuard>
            </ModalsProvider>
        </MantineProvider>
    );
};

export default App;
