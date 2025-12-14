export const safeJsonParser = <T = any>(jsonString: string): T | null => {
    try {
        return JSON.parse(jsonString) as T;
    } catch {
        return null;
    }
};
