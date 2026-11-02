import { create } from "zustand";
import { ResponseOK, Track, TrackKind } from "../api/types";
import { airstationAPI } from "../api";

interface TracksStore {
    tracks: Track[];
    totalTracks: number;

    setTracks(tracks: Track[]): void;
    fetchTracks(
        p: number,
        l: number,
        s: string,
        sb: keyof Track,
        so: "asc" | "desc",
        kind?: TrackKind | "all",
    ): Promise<void>;
    uploadTracks(files: File[], kind?: TrackKind): Promise<ResponseOK>;
    deleteTracks(trackIDs: string[]): Promise<ResponseOK>;
}

export const useTracksStore = create<TracksStore>()((set, get) => ({
    tracks: [],
    totalTracks: 0,

    setTracks(q) {
        set({ tracks: q });
    },

    async fetchTracks(p, l, s, sb, so, kind = "all") {
        const result = await airstationAPI.getTracks(p, l, s, sb, so, kind);
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

    async uploadTracks(files, kind = "music") {
        const resp = await airstationAPI.uploadTracks(files, kind);
        return resp;
    },

    async deleteTracks(trackIDs) {
        const resp = await airstationAPI.deleteTracks(trackIDs);
        const filtered = get().tracks.filter(({ id }) => !trackIDs.includes(id));
        set({ tracks: filtered, totalTracks: get().totalTracks - trackIDs.length });
        return resp;
    },
}));
