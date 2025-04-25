import { create } from "zustand";
import { ResponseOK, Track } from "../api/types";
import { airstationAPI } from "../api";

interface TracksStore {
    tracks: Track[];

    setTracks(tracks: Track[]): void;
    fetchTracks(p: number, l: number, s: string): Promise<void>;
    uploadTracks(files: File[]): Promise<ResponseOK>;
    deleteTracks(trackIDs: string[]): Promise<ResponseOK>;
}

export const useTracksStore = create<TracksStore>()((set, get) => ({
    tracks: [],

    setTracks(q) {
        set({ tracks: q });
    },

    async fetchTracks(p: number, l: number, s: string) {
        const { tracks } = await airstationAPI.getTracks(p, l, s);
        set({ tracks });
    },

    async uploadTracks(files: File[]) {
        const resp = await airstationAPI.uploadTracks(files);
        return resp;
    },

    async deleteTracks(trackIDs: string[]) {
        const resp = await airstationAPI.deleteTracks(trackIDs);
        const filtered = get().tracks.filter(({ id }) => !trackIDs.includes(id));
        set({ tracks: filtered });
        return resp;
    },
}));
