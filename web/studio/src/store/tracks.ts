import { create } from "zustand";
import { Track } from "../api/types";
import { airstationAPI } from "../api";
import { errNotify } from "../notifications";

interface TracksStore {
    tracks: Track[];

    fetchTracks: (p: number, l: number, s: string) => Promise<void>;
    setTracks: (tracks: Track[]) => void;
    addTracks: (tracks: Track[]) => void;
}

export const useTracksStore = create<TracksStore>()((set) => ({
    tracks: [],

    async fetchTracks(p: number, l: number, s: string) {
        try {
            const { tracks } = await airstationAPI.getTracks(p, l, s);
            set({ tracks });
        } catch (error) {
            errNotify(error);
        }
    },

    setTracks(q) {
        set({ tracks: q });
    },

    addTracks(tracks) {
        set((state) => ({ tracks: [...tracks, ...state.tracks] }));
    },
}));
