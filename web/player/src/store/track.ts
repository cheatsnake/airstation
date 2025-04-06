import { createStore } from "solid-js/store";

export const [trackStore, setTrackStore] = createStore({
    trackName: "",
    isPlay: false,
});
