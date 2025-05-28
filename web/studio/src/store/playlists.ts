import { create } from "zustand";
import { Playlist, ResponseOK } from "../api/types";
import { airstationAPI } from "../api";

interface PlaylistStore {
    playlists: Playlist[];

    setPlaylists(playlists: Playlist[]): void;
    addPlaylist(name: string, trackIDs: string[], description?: string): Promise<Playlist>;
    fetchPlaylists(): Promise<void>;
    editPlaylist(modified: Playlist): Promise<ResponseOK>;
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

    async editPlaylist(modified) {
        const resp = await airstationAPI.editPlaylist(
            modified.id,
            modified.name,
            modified.tracks.map(({ id }) => id),
            modified.description,
        );

        return resp;
    },

    async deletePlaylist(id) {
        const resp = await airstationAPI.deletePlaylist(id);
        return resp;
    },
}));
