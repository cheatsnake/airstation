// Feature detection for native HLS playback. Returns true on Safari (iOS/macOS),
// false on Chrome/Firefox/Edge. Prefer this over UA sniffing for source selection.
export const canPlayNativeHls = (el: HTMLMediaElement | null | undefined): boolean =>
    !!el && !!el.canPlayType("application/vnd.apple.mpegurl");
