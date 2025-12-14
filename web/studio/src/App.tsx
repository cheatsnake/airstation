import { MantineProvider } from "@mantine/core";
import { ModalsProvider } from "@mantine/modals";
import { Notifications } from "@mantine/notifications";
import { AuthGuard } from "./components/AuthGuard";
import { theme } from "./theme";
import { Page } from "./page";

const App = () => {
    return (
        <MantineProvider defaultColorScheme="dark" theme={theme}>
            <ModalsProvider modalProps={{ transitionProps: { duration: 100 } }}>
                <Notifications position="bottom-right" autoClose={7000} />
                <AuthGuard>
                    <Page />
                </AuthGuard>
            </ModalsProvider>
        </MantineProvider>
    );
};

export default App;
