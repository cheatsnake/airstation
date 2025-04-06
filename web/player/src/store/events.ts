import { createStore } from "solid-js/store";
import { API_HOST, API_PREFIX } from "../api";

export const EVENT_SOURCE_URL = API_HOST + API_PREFIX + "/events";
export const EVENTS = {
    newTrack: "new_track",
    countListeners: "count_listeners",
    pause: "pause",
    play: "play",
};

const [eventSourceStore, setEventSourceStore] = createStore<{ eventSource: EventSource | null }>({
    eventSource: null,
});

export const initEventSource = () => {
    if (eventSourceStore.eventSource) eventSourceStore.eventSource.close();

    const es = new EventSource(EVENT_SOURCE_URL);
    setEventSourceStore("eventSource", es);
};

export const addEventListener = (event: string, listener: (event: MessageEvent) => void) => {
    if (eventSourceStore.eventSource) {
        eventSourceStore.eventSource.addEventListener(event, listener);
    }
};

export const closeEventSource = () => {
    if (eventSourceStore.eventSource) {
        eventSourceStore.eventSource.close();
        setEventSourceStore("eventSource", null);
    }
};
