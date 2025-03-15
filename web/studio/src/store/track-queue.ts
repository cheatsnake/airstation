import { create } from "zustand";
import { Track } from "../api/types";
import { airstationAPI } from "../api";

interface TrackQueueStore {
    queue: Track[];
    fetchQueue: () => Promise<void>;
    updateQueue: (tracks: Track[]) => Promise<void>;
}

export const useTrackQueueStore = create<TrackQueueStore>()((set) => ({
    queue: [],

    async fetchQueue() {
        const q = await airstationAPI.getQueue();
        set({ queue: q });
    },

    async updateQueue(tracks) {
        await airstationAPI.updateQueue(tracks.map(({ id }) => id));
        set({ queue: tracks });
    },
}));
