import { createStore } from "solid-js/store";

export const [trackStore, setTrackStore] = createStore({
    trackName: "",
    nextTrackName: "",
    voiceOverName: "",
    isPlay: false,
});
