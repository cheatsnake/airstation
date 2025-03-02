export const handleErr = (err: unknown) => {
    return String(err).replace("Error: ", "");
};
