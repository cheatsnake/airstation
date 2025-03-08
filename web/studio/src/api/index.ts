import { PlaybackState, ResponseErr, ResponseOK, Track, TracksPage } from "./types";
import { jsonRequestParams, queryParams } from "./utils";

export const API_HOST = "";
const API_PREFIX = "/v1/api";

class AirstationAPI {
    private host: string;
    private prefix: string;
    private url: () => string;

    constructor(host: string, prefix: string) {
        this.host = host;
        this.prefix = prefix;
        this.url = () => `${this.host + this.prefix}`;
    }

    async login(secret: string) {
        const url = `${this.url()}/login`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("POST", { secret }));
    }

    async getPlayback() {
        const url = `${this.url()}/playback`;
        return await this.makeRequest<PlaybackState>(url);
    }

    async getTracks(page: number, limit: number, search: string) {
        const url = `${this.url()}/tracks?${queryParams({ page, limit, search })}`;
        return await this.makeRequest<TracksPage>(url);
    }

    async uploadTracks(files: File[]) {
        const url = `${this.url()}/tracks`;
        const formData = new FormData();

        for (let i = 0; i < files.length; i++) {
            formData.append("tracks", files[i]);
        }

        return await this.makeRequest<Track[]>(url, {
            method: "POST",
            body: formData,
        });
    }

    async deleteTracks(ids: string[]) {
        const url = `${this.url()}/tracks`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("DELETE", { ids }));
    }

    async getQueue() {
        const url = `${this.url()}/queue`;
        return await this.makeRequest<Track[]>(url);
    }

    async addToQueue(trackIDs: string[]) {
        const url = `${this.url}/queue`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("POST", { ids: trackIDs }));
    }

    async updateQueue(trackIDs: string[]) {
        const url = `${this.url}/queue`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("PUT", { ids: trackIDs }));
    }

    async removeFromQueue(trackIDs: string[]) {
        const url = `${this.url}/queue`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("DELETE", { ids: trackIDs }));
    }

    private async makeRequest<T>(url: string, params: RequestInit = {}): Promise<T> {
        params.headers = { ...params.headers };

        const resp = await fetch(url, params);
        if (!resp.ok) {
            const body: ResponseErr = await resp.json();
            throw new Error(body.message);
        }

        return resp.json();
    }
}

export const airstationAPI = new AirstationAPI(API_HOST, API_PREFIX);
