import { create } from "zustand";
import { PlaybackState } from "../api/types";
import { airstationAPI } from "../api";
import { errNotify } from "../notifications";

interface PlaybackStore {
    playback: PlaybackState;
    fetchPlayback: () => Promise<void>;
    incElapsedTime: (value: number) => void;
}

export const usePlaybackStore = create<PlaybackStore>()((set) => ({
    playback: { currentTrack: null, currentTrackElapsed: 0, isPlaying: false },

    async fetchPlayback() {
        try {
            const pb = await airstationAPI.getPlayback();
            set({ playback: pb });
        } catch (error) {
            errNotify(error);
        }
    },

    incElapsedTime(value) {
        set((state) => {
            if (!state.playback.currentTrack) return state;

            const elapsed = state.playback.currentTrackElapsed + value;
            if (elapsed > state.playback.currentTrack.duration) return state;

            return {
                ...state,
                playback: {
                    ...state.playback,
                    currentTrackElapsed: elapsed,
                },
            };
        });
    },
}));
