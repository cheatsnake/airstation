import { createSignal } from "solid-js";
import { PlaybackHistory } from "../api/types";

export const [history, setHistory] = createSignal<PlaybackHistory[]>([]);
export const addHistory = (h: PlaybackHistory) => {
    setHistory([h, ...history()]);
};
