export const formatDateToTimeFirst = (date: Date) => {
    const timeParts = new Intl.DateTimeFormat(undefined, {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false,
    }).formatToParts(date);

    const dateParts = new Intl.DateTimeFormat(undefined, {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
    }).formatToParts(date);

    const time = timeParts.map((p) => p.value).join("");
    const dateStr = dateParts.map((p) => p.value).join("");

    return `${time} ${dateStr}`;
};

export const getUnixTime = (): number => Math.floor(Date.now() / 1000);
