import { create } from "zustand";
import { persist } from "zustand/middleware";

interface SettingsStore {
    interfaceWidth?: number;

    setInterfaceWidth: (width: number) => void;
}

export const useSettingsStore = create<SettingsStore>()(
    persist(
        (set) => ({
            setInterfaceWidth: (width: number) => set({ interfaceWidth: width }),
        }),
        { name: "settings" },
    ),
);
