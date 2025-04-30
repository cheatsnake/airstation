import { create } from "zustand";
import { ResponseOK, Track } from "../api/types";
import { airstationAPI } from "../api";

interface TracksStore {
    tracks: Track[];
    totalTracks: number;

    setTracks(tracks: Track[]): void;
    fetchTracks(p: number, l: number, s: string, sb: keyof Track, so: "asc" | "desc"): Promise<void>;
    uploadTracks(files: File[]): Promise<ResponseOK>;
    deleteTracks(trackIDs: string[]): Promise<ResponseOK>;
}

export const useTracksStore = create<TracksStore>()((set, get) => ({
    tracks: [],
    totalTracks: 0,

    setTracks(q) {
        set({ tracks: q });
    },

    async fetchTracks(p: number, l: number, s: string, sb: keyof Track, so: "asc" | "desc") {
        const result = await airstationAPI.getTracks(p, l, s, sb, so);
        if (p === 1) {
            set({ tracks: result.tracks, totalTracks: result.total });
            return;
        }

        // If it not a first page, just append new tracks
        const trackIDs = new Set(get().tracks.map((t) => t.id));
        const tracks = [...get().tracks];

        for (const track of result.tracks) {
            if (trackIDs.has(track.id)) continue;
            tracks.push(track);
        }

        set({ tracks, totalTracks: result.total });
    },

    async uploadTracks(files: File[]) {
        const resp = await airstationAPI.uploadTracks(files);
        return resp;
    },

    async deleteTracks(trackIDs: string[]) {
        const resp = await airstationAPI.deleteTracks(trackIDs);
        const filtered = get().tracks.filter(({ id }) => !trackIDs.includes(id));
        set({ tracks: filtered, totalTracks: get().totalTracks - trackIDs.length });
        return resp;
    },
}));
