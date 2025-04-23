import { create } from "zustand";
import { PlaybackState } from "../api/types";
import { airstationAPI } from "../api";
import { errNotify } from "../notifications";
import { getUnixTime } from "../utils/time";

interface PlaybackStore {
    playback: PlaybackState;
    setPlayback: (pb: PlaybackState) => void;
    fetchPlayback: () => Promise<void>;
    syncElapsedTime: () => void;
}

export const usePlaybackStore = create<PlaybackStore>()((set) => ({
    playback: { currentTrack: null, currentTrackElapsed: 0, isPlaying: false, updatedAt: getUnixTime() },

    setPlayback(pb) {
        if (pb.currentTrack) pb.currentTrack.duration = Math.ceil(pb.currentTrack.duration);
        set({ playback: pb });
    },

    async fetchPlayback() {
        try {
            const pb = await airstationAPI.getPlayback();
            if (pb.currentTrack) pb.currentTrack.duration = Math.ceil(pb.currentTrack.duration);

            set({ playback: pb });
        } catch (error) {
            errNotify(error);
        }
    },

    syncElapsedTime() {
        set((state) => {
            if (!state.playback.currentTrack) return state;

            const currentTime = getUnixTime();
            const diff = currentTime - state.playback.updatedAt;
            const elapsed = state.playback.currentTrackElapsed + diff;
            if (elapsed > state.playback.currentTrack.duration) return state;

            return {
                ...state,
                playback: {
                    ...state.playback,
                    currentTrackElapsed: elapsed,
                    updatedAt: currentTime,
                },
            };
        });
    },
}));
