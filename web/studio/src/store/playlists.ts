import { create } from "zustand";
import { Playlist, ResponseOK } from "../api/types";
import { airstationAPI } from "../api";

interface PlaylistStore {
    playlists: Playlist[];

    setPlaylists(playlists: Playlist[]): void;
    addPlaylist(name: string, trackIDs: string[], description?: string): Promise<Playlist>;
    fetchPlaylists(): Promise<void>;
    editPlaylist(id: string, name: string, trackIDs: string[], description?: string): Promise<ResponseOK>;
    deletePlaylist(id: string): Promise<ResponseOK>;
}

export const usePlaylistStore = create<PlaylistStore>()((set, get) => ({
    playlists: [],

    setPlaylists(p) {
        set({ playlists: p });
    },

    async fetchPlaylists() {
        const p = await airstationAPI.getPlaylists();
        set({ playlists: p });
    },

    async addPlaylist(name, trackIDs, description) {
        const p = await airstationAPI.addPlaylist(name, trackIDs, description);
        set({ playlists: [p, ...get().playlists] });
        return p;
    },

    async editPlaylist(id: string, name: string, trackIDs: string[], description?: string) {
        const resp = await airstationAPI.editPlaylist(id, name, trackIDs, description);
        set({
            playlists: get().playlists.map((p) =>
                p.id === id
                    ? {
                          id,
                          name,
                          tracks: [],
                          trackCount: trackIDs.length,
                          description,
                      }
                    : p,
            ),
        });

        return resp;
    },

    async deletePlaylist(id) {
        const resp = await airstationAPI.deletePlaylist(id);
        set({ playlists: get().playlists.filter((p) => p.id !== id) });
        return resp;
    },
}));
