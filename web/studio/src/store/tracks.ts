import { create } from "zustand";
import { Track } from "../api/types";

interface TracksStore {
    tracks: Track[];
    setTracks: (tracks: Track[]) => void;
}

export const useTracksStore = create<TracksStore>()((set) => ({
    tracks: [],

    setTracks(q) {
        set({ tracks: q });
    },
}));
