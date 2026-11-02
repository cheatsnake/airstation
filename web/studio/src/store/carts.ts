import { create } from "zustand";
import { airstationAPI } from "../api";
import { ResponseOK, Track } from "../api/types";

interface CartsStore {
    carts: Track[];
    fetchCarts: () => Promise<void>;
    playCart: (id: string, mode: "next" | "now") => Promise<ResponseOK>;
    deleteCart: (id: string) => Promise<ResponseOK>;
}

export const useCartsStore = create<CartsStore>()((set, get) => ({
    carts: [],

    async fetchCarts() {
        const carts = await airstationAPI.getCarts();
        set({ carts });
    },

    async playCart(id, mode) {
        return await airstationAPI.playCart(id, mode);
    },

    async deleteCart(id) {
        const resp = await airstationAPI.deleteTracks([id]);
        set({ carts: get().carts.filter((cart) => cart.id !== id) });
        return resp;
    },
}));
