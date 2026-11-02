import { create } from "zustand";
import { airstationAPI } from "../api";
import { VoiceOver, VoiceOverDraft } from "../api/types";

interface VoiceOverStore {
    voiceOvers: VoiceOver[];
    fetchVoiceOvers: () => Promise<void>;
    createVoiceOver: (draft: VoiceOverDraft) => Promise<VoiceOver>;
    cancelVoiceOver: (id: string) => Promise<void>;
}

export const useVoiceOverStore = create<VoiceOverStore>()((set, get) => ({
    voiceOvers: [],

    async fetchVoiceOvers() {
        const voiceOvers = await airstationAPI.getVoiceOvers();
        set({ voiceOvers });
    },

    async createVoiceOver(draft) {
        const created = await airstationAPI.createVoiceOver(draft);
        set({ voiceOvers: [created, ...get().voiceOvers] });
        return created;
    },

    async cancelVoiceOver(id) {
        await airstationAPI.cancelVoiceOver(id);
        set({
            voiceOvers: get().voiceOvers.map((vo) =>
                vo.id === id ? { ...vo, status: "cancelled" as const } : vo,
            ),
        });
    },
}));
