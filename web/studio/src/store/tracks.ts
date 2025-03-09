import { create } from "zustand";
import { Track } from "../api/types";
import { airstationAPI } from "../api";

interface TracksStore {
    tracks: Track[];

    fetchTracks: (p: number, l: number, s: string) => Promise<void>;
    setTracks: (tracks: Track[]) => void;
    addTracks: (tracks: Track[]) => void;
}

export const useTracksStore = create<TracksStore>()((set) => ({
    tracks: [],

    async fetchTracks(p: number, l: number, s: string) {
        const { tracks } = await airstationAPI.getTracks(p, l, s);
        set({ tracks });
    },

    setTracks(q) {
        set({ tracks: q });
    },

    addTracks(tracks) {
        set((state) => ({ tracks: [...tracks, ...state.tracks] }));
    },
}));
