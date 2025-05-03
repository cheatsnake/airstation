import { FC, JSX, useEffect, useState } from "react";
import { airstationAPI } from "../api";
import { useDisclosure } from "@mantine/hooks";
import { errNotify } from "../notifications";
import { handleErr } from "../utils/error";
import { Box, Button, Flex, Group, LoadingOverlay, Paper, TextInput } from "@mantine/core";

export const AuthGuard: FC<{ children: JSX.Element }> = (props) => {
    const [isAuth, setIsAuth] = useState(false);
    const [loader, handLoader] = useDisclosure(false);

    const handleLogin = async (secret: string) => {
        try {
            handLoader.open();
            await airstationAPI.login(secret);
            setIsAuth(true);
        } catch (error) {
            errNotify(error);
        } finally {
            handLoader.close();
        }
    };

    useEffect(() => {
        (async () => {
            try {
                handLoader.open();
                await airstationAPI.getQueue();
                setIsAuth(true);
            } catch (error) {
                const msg = handleErr(error);
                if (!msg.includes("Unauthorized")) errNotify(msg);
            } finally {
                handLoader.close();
            }
        })();
    }, []);

    return (
        <>
            {isAuth ? (
                props.children
            ) : (
                <Box w="100%" h="100vh">
                    <LoadingOverlay visible={loader} />
                    {loader ? null : <LoginForm handleLogin={handleLogin} />}
                </Box>
            )}
        </>
    );
};

const LoginForm: FC<{ handleLogin: (s: string) => Promise<void> }> = (props) => {
    const [secret, setSecret] = useState("");

    return (
        <Flex h="100%" justify="center" align="center">
            <Paper mb="sm" w={250}>
                <TextInput
                    type="password"
                    required
                    value={secret}
                    onChange={(event) => setSecret(event.currentTarget.value)}
                    placeholder="Enter secret"
                />
                <Group mt="sm" justify="center">
                    <Button
                        disabled={secret.length < 10}
                        fullWidth
                        variant="light"
                        onClick={() => props.handleLogin(secret)}
                    >
                        Submit
                    </Button>
                </Group>
            </Paper>
        </Flex>
    );
};
