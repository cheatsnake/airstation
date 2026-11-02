import { create } from "zustand";
import { airstationAPI } from "../api";
import { ResponseOK, ScheduleRule, ScheduleRuleDraft } from "../api/types";

interface ScheduleStore {
    rules: ScheduleRule[];
    fetchRules: () => Promise<void>;
    createRule: (draft: ScheduleRuleDraft) => Promise<ScheduleRule>;
    editRule: (id: string, draft: ScheduleRuleDraft) => Promise<ScheduleRule>;
    deleteRule: (id: string) => Promise<ResponseOK>;
}

export const useScheduleStore = create<ScheduleStore>()((set, get) => ({
    rules: [],

    async fetchRules() {
        const rules = await airstationAPI.getScheduleRules();
        set({ rules });
    },

    async createRule(draft) {
        const created = await airstationAPI.createScheduleRule(draft);
        set({ rules: [...get().rules, created] });
        return created;
    },

    async editRule(id, draft) {
        const updated = await airstationAPI.editScheduleRule(id, draft);
        set({ rules: get().rules.map((r) => (r.id === id ? updated : r)) });
        return updated;
    },

    async deleteRule(id) {
        const resp = await airstationAPI.deleteScheduleRule(id);
        set({ rules: get().rules.filter((r) => r.id !== id) });
        return resp;
    },
}));
