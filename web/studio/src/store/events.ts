import { create } from "zustand";
import { API_HOST, API_PREFIX } from "../api";

const EVENT_SOURCE_URL = API_HOST + API_PREFIX + "/events";
export const EVENTS = {
    newTrack: "new_track",
    loadedTracks: "loaded_tracks",
    countListeners: "count_listeners",
};

type EventHandler = (event: MessageEvent) => void;

interface EventSourceStore {
    eventSource?: EventSource;
    addEventHandler: (eventName: string, handler: EventHandler) => void;
    closeEventSource: () => void;
}

export const useEventSourceStore = create<EventSourceStore>((set, get) => ({
    eventSource: undefined,

    addEventHandler: (eventName: string, handler: EventHandler) => {
        let { eventSource } = get();

        if (!eventSource) {
            eventSource = new EventSource(EVENT_SOURCE_URL);

            eventSource.onerror = () => {
                console.error("EventSource connection error");
            };

            set({ eventSource });
        }

        eventSource.addEventListener(eventName, handler);
    },

    closeEventSource: () => {
        const { eventSource } = get();
        if (eventSource) {
            eventSource.close();
            set({ eventSource: undefined });
        }
    },
}));
