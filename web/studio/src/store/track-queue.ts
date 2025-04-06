import { create } from "zustand";
import { ResponseOK, Track } from "../api/types";
import { airstationAPI } from "../api";

interface TrackQueueStore {
    queue: Track[];
    fetchQueue: () => Promise<void>;
    updateQueue: (tracks: Track[]) => Promise<void>;
    addToQueue(trackIDs: string[]): Promise<ResponseOK>;
    removeFromQueue(trackIDs: string[]): Promise<ResponseOK>;
    rotateQueue: () => void;
}

export const useTrackQueueStore = create<TrackQueueStore>()((set, get) => ({
    queue: [],

    async fetchQueue() {
        const q = await airstationAPI.getQueue();
        set({ queue: q });
    },

    async updateQueue(tracks) {
        set({ queue: tracks });
        await airstationAPI.updateQueue(tracks.map(({ id }) => id));
    },

    async addToQueue(trackIDs: string[]) {
        const resp = await airstationAPI.addToQueue(trackIDs);
        const q = await airstationAPI.getQueue();
        set({ queue: q });
        return resp;
    },

    async removeFromQueue(trackIDs: string[]) {
        const resp = await airstationAPI.removeFromQueue(trackIDs);
        const filtered = get().queue.filter(({ id }) => !trackIDs.includes(id));
        set({ queue: filtered });
        return resp;
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
