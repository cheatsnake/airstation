import { create } from "zustand";
import { PlaybackState } from "../api/types";

interface PlaybackStore {
  playback: PlaybackState | null;
  setPlayback: (pb: PlaybackState) => void;
}

export const usePlaybackStore = create<PlaybackStore>()((set) => ({
  playback: null,

  setPlayback(pb) {
    set({ playback: pb });
  },
}));
