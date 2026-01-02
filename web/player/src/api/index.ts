import { PlaybackHistory, PlaybackState, ResponseErr, StationInfo } from "./types";
import { queryParams } from "./utils";

export const API_HOST = "";
export const API_PREFIX = "/api/v1";

class AirstationAPI {
    private host: string;
    private prefix: string;
    private url: () => string;

    constructor(host: string, prefix: string) {
        this.host = host;
        this.prefix = prefix;
        this.url = () => `${this.host + this.prefix}`;
    }

    async getPlayback() {
        const url = `${this.url()}/playback`;
        return await this.makeRequest<PlaybackState>(url);
    }

    async getPlaybackHistory(limit?: number) {
        let url = `${this.url()}/playback/history`;
        if (limit) url += `?${queryParams({ limit })}`;
        return await this.makeRequest<PlaybackHistory[]>(url);
    }

    async getStationInfo() {
        const url = `${this.url()}/station/info`;
        return await this.makeRequest<StationInfo>(url);
    }

    private async makeRequest<T>(url: string, params: RequestInit = {}): Promise<T> {
        const resp = await fetch(url, params);
        if (!resp.ok) {
            const body: ResponseErr = await resp.json();
            throw new Error(body.message);
        }

        return resp.json();
    }
}

export const airstationAPI = new AirstationAPI(API_HOST, API_PREFIX);
