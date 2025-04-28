export const moveArrayItem = <T>(array: T[], fromIndex: number, toIndex: number): T[] => {
    if (fromIndex < 0 || fromIndex >= array.length || toIndex < 0 || toIndex >= array.length) {
        return array;
    }

    const newArray = [...array];
    const [movedItem] = newArray.splice(fromIndex, 1);

    newArray.splice(toIndex, 0, movedItem);
    return newArray;
};

export const shuffleArray = <T>(array: T[]): T[] => {
    const shuffled = [...array];

    for (let i = shuffled.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
    }

    return shuffled;
};
