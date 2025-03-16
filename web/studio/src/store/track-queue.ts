import { create } from "zustand";
import { Track } from "../api/types";
import { airstationAPI } from "../api";
import { errNotify } from "../notifications";

interface TrackQueueStore {
    queue: Track[];
    fetchQueue: () => Promise<void>;
    updateQueue: (tracks: Track[]) => Promise<void>;
    rotateQueue: () => void;
}

export const useTrackQueueStore = create<TrackQueueStore>()((set) => ({
    queue: [],

    async fetchQueue() {
        try {
            const q = await airstationAPI.getQueue();
            set({ queue: q });
        } catch (error) {
            errNotify(error);
        }
    },

    async updateQueue(tracks) {
        set({ queue: tracks });
        await airstationAPI.updateQueue(tracks.map(({ id }) => id));
    },

    rotateQueue() {
        set((state) => {
            if (state.queue.length === 0) return state;
            return {
                ...state,
                queue: [...state.queue.slice(1), state.queue[0]],
            };
        });
    },
}));
