import { createSignal } from "solid-js";

const STORAGE_KEY = "airstation:volume";
const DEFAULT_VOLUME = 1;

const clampVolume = (value: number) => Math.min(1, Math.max(0, value));

const loadVolume = (): number => {
    try {
        const stored = window.localStorage.getItem(STORAGE_KEY);
        if (stored === null) return DEFAULT_VOLUME;

        const value = Number.parseFloat(stored);
        return Number.isFinite(value) ? clampVolume(value) : DEFAULT_VOLUME;
    } catch {
        return DEFAULT_VOLUME;
    }
};

const [volume, setVolumeSignal] = createSignal(loadVolume());

export { volume };

export const setVolume = (value: number) => {
    const next = clampVolume(value);
    setVolumeSignal(next);

    try {
        window.localStorage.setItem(STORAGE_KEY, String(next));
    } catch {
        // Storage can be unavailable (e.g. private browsing); keep the in-memory value.
    }
};
