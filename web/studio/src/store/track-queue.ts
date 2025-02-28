import { create } from "zustand";
import { Track } from "../api/types";

interface TrackQueueStore {
    queue: Track[];
    setQueue: (tracks: Track[]) => void;
}

export const useTrackQueueStore = create<TrackQueueStore>()((set) => ({
    queue: [],

    setQueue(q) {
        set({ queue: q });
    },
}));
