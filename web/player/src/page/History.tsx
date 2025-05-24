import { Accessor, Component, createSignal, onMount } from "solid-js";
import { airstationAPI } from "../api";
import styles from "./History.module.css";
import { formatDateToTimeFirst } from "../utils/date";
import { history, setHistory } from "../store/history";

export const History = () => {
    const [isOpen, setIsOpen] = createSignal(false);
    const open = () => setIsOpen(true);
    const close = () => setIsOpen(false);

    return (
        <>
            <div
                tabIndex={0}
                role="button"
                class={`${isOpen() ? "empty_icon" : styles.menu_icon}`}
                onClick={open}
            ></div>
            <Menu isOpen={isOpen} close={close} />
        </>
    );
};

const DESKTOP_WIDTH = 1100;
const MAX_HISTORY_LIMIT = 500;

const Menu: Component<{ isOpen: Accessor<boolean>; close: () => void }> = ({ isOpen, close }) => {
    const [hideLoadMore, setHideLoadMore] = createSignal(false);
    const loadHistory = async (limit?: number) => {
        try {
            const h = await airstationAPI.getPlaybackHistory(limit);
            setHistory(h);
        } catch (error) {
            console.log(error);
        }
    };

    const loadMoreHistory = () => {
        loadHistory(MAX_HISTORY_LIMIT);
        setHideLoadMore(true);
    };

    const copyToClipboard = async (text: string) => {
        try {
            await navigator.clipboard.writeText(text);
        } catch (error) {
            console.log(error);
        }
    };

    onMount(() => {
        loadHistory();
    });

    return (
        <div
            class={`${styles.menu} ${isOpen() ? styles.open : ""} ${
                window.screen.width > DESKTOP_WIDTH ? styles.menu_desktop : styles.menu_mobile
            }`}
        >
            <div class={styles.menu_header}>
                <div class={styles.menu_title}></div>
                <div tabIndex={0} role="button" class={styles.close_icon} onClick={close}></div>
            </div>
            <div class={styles.history}>
                {history().map((h) => (
                    <div class={styles.history_item} onClick={() => copyToClipboard(h.trackName)}>
                        <div class={styles.history_name}>{h.trackName}</div>
                        <div class={styles.history_timestamp}>{formatDateToTimeFirst(new Date(h.playedAt * 1000))}</div>
                    </div>
                ))}
                {hideLoadMore() ? null : (
                    <button class={styles.load_more_btn} onClick={loadMoreHistory}>
                        Load more
                    </button>
                )}
            </div>
        </div>
    );
};
