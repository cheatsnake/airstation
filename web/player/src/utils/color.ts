export function getHueFromHex(hex: string) {
    hex = hex.replace("#", "");

    const r = parseInt(hex.slice(0, 2), 16) / 255;
    const g = parseInt(hex.slice(2, 4), 16) / 255;
    const b = parseInt(hex.slice(4, 6), 16) / 255;

    const max = Math.max(r, g, b);
    const min = Math.min(r, g, b);
    const d = max - min;

    if (d === 0) return 0;

    let h;
    switch (max) {
        case r:
            h = ((g - b) / d) % 6;
            break;
        case g:
            h = (b - r) / d + 2;
            break;
        default:
            h = (r - g) / d + 4;
    }

    return Math.round(h * 60 < 0 ? h * 60 + 360 : h * 60);
}
