export const defineBoxHeight = (windowHeight: number) => {
    if (windowHeight > 1600) return 1600 * 0.6;
    return windowHeight * 0.6;
};
