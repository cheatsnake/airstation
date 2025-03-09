import { create } from "zustand";
import { Track } from "../api/types";
import { airstationAPI } from "../api";

interface TrackQueueStore {
    queue: Track[];
    fetchQueue: () => Promise<void>;
    setQueue: (tracks: Track[]) => void;
}

export const useTrackQueueStore = create<TrackQueueStore>()((set) => ({
    queue: [],

    async fetchQueue() {
        const q = await airstationAPI.getQueue();
        set({ queue: q });
    },
    setQueue(q) {
        set({ queue: q });
    },
}));
