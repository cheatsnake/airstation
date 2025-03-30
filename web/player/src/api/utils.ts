export const jsonRequestParams = (method: string, body: Record<string, any>) => {
    return {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    };
};

export const queryParams = (params: Record<string, any>) => {
    removeEmptyFields(params);
    return new URLSearchParams(params).toString();
};

const removeEmptyFields = (obj: Record<string, any>) => {
    for (const key in obj) {
        if (obj.hasOwnProperty(key) && [undefined, null, ""].includes(obj[key])) {
            delete obj[key];
        }
    }
};
