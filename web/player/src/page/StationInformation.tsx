import { Accessor, Component, createSignal, onMount } from "solid-js";
import pageStyles from "./Page.module.css";
import styles from "./StationInformation.module.css";
import { airstationAPI } from "../api";
import { DESKTOP_WIDTH } from "../const";
import { StationInfo } from "../api/types";
import { isValidURL } from "../utils/url";
import { isValidHexColor } from "../utils/color";
import { setCssVariable, setFavicon, setPageTitle } from "../utils/document";
import { addEventListener, EVENTS } from "../store/events";

export const StationInformation = () => {
    const [isOpen, setIsOpen] = createSignal(false);
    const open = () => setIsOpen(true);
    const close = () => setIsOpen(false);

    return (
        <>
            <div role="button" class={`${isOpen() ? "empty_icon" : styles.info_icon}`} onClick={open} />
            <Card isOpen={isOpen} close={close} />
        </>
    );
};

const parseLinks = (rawLinks: string): { title: string; url: string }[] => {
    const regex = /\[([^\]]+)]\((https?:\/\/[^\s)]+)\)/g;
    return Array.from(rawLinks.matchAll(regex), (m) => ({
        title: m[1],
        url: m[2],
    }));
};

const parseTheme = (rawTheme: string) => {
    const [bgStart, bgEnd, bgIcon, text, accent, bgImage] = rawTheme.split(";");

    if (bgStart && isValidHexColor(bgStart)) setCssVariable("--bg-gradient-start", bgStart);
    if (bgEnd && isValidHexColor(bgEnd)) setCssVariable("--bg-gradient-end", bgEnd);
    if (bgIcon && isValidHexColor(bgIcon)) setCssVariable("--bg-icon", bgIcon);
    if (text && isValidHexColor(text)) setCssVariable("--text-color", text);

    if (accent && isValidHexColor(accent)) {
        setCssVariable("--accent-color", accent);
    } else {
        setCssVariable("--accent-color", "");
    }

    if (bgImage && isValidURL(bgImage)) {
        document.body.style.backgroundImage = `url(${bgImage})`;
    } else {
        document.body.style.backgroundImage = "";
    }
};

const Card: Component<{ isOpen: Accessor<boolean>; close: () => void }> = ({ isOpen, close }) => {
    const [info, setInfo] = createSignal<StationInfo | null>(null);

    const loadInfo = async () => {
        try {
            const h = await airstationAPI.getStationInfo();
            setInfo(h);
            if (h.name) setPageTitle(h.name);
            if (isValidURL(h.faviconURL)) setFavicon(h.faviconURL);
            if (h.theme) parseTheme(h.theme);
        } catch (error) {
            console.log(error);
        }
    };

    onMount(() => {
        loadInfo();

        addEventListener(EVENTS.changeTheme, (_e: MessageEvent<string>) => {
            loadInfo();
        });
    });

    return (
        <div
            class={`${styles.info_menu} ${isOpen() ? styles.info_open : ""} ${
                window.screen.width > DESKTOP_WIDTH ? pageStyles.menu_desktop : pageStyles.menu_mobile
            }`}
        >
            <div class={styles.header}>
                <div role="button" class={pageStyles.close_icon} onClick={close}></div>
            </div>

            {info()?.logoURL && <img src={info()?.logoURL} alt={info?.name} class={styles.logo} />}

            <div class={styles.content}>
                <div class={styles.title}>{info()?.name}</div>

                <div class={styles.metadata}>
                    {info()?.location && <span class={styles.location}>{info()!.location}</span>}
                    {info()?.timezone && <span class={styles.timezone}>{info()!.timezone}</span>}
                </div>

                <div class={styles.description} innerHTML={info()?.description} />

                {info()?.links && (
                    <div class={styles.footer}>
                        {parseLinks(info()?.links!).map((link) => (
                            <a href={link.url} target="_blank" rel="noreferrer">
                                {link.title}
                            </a>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};
