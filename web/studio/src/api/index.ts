import { PlaybackState, ResponseErr, ResponseOK, Track, TracksPage } from "./types";
import { jsonRequestParams, queryParams } from "./utils";

export const API_HOST = "";
const API_PREFIX = "/v1/api";
const AUTH_TOKEN = "test111111";

class AirstationAPI {
    private host: string;
    private prefix: string;
    private authToken: string;
    private url: () => string;

    constructor(host: string, prefix: string, authToken: string) {
        this.host = host;
        this.prefix = prefix;
        this.authToken = authToken;
        this.url = () => `${this.host + this.prefix}`;
    }

    async getPlayback() {
        const url = `${this.url()}/playback`;
        return await this.fetchWithAuth<PlaybackState>(url);
    }

    async getTracks(page: number, limit: number, search: string) {
        const url = `${this.url()}/tracks?${queryParams({ page, limit, search })}`;
        return await this.fetchWithAuth<TracksPage>(url);
    }

    async uploadTracks(files: FileList) {
        const url = `${this.url()}/tracks`;
        const formData = new FormData();

        for (let i = 0; i < files.length; i++) {
            formData.append("files", files[i]);
        }

        return await this.fetchWithAuth<Track[]>(url, {
            method: "POST",
            body: formData,
        });
    }

    async deleteTracks(trackIDs: string[]) {
        const url = `${this.url()}/tracks`;
        return await this.fetchWithAuth<ResponseOK>(url, jsonRequestParams("DELETE", { ids: trackIDs }));
    }

    async getQueue() {
        const url = `${this.url()}/queue`;
        return await this.fetchWithAuth<Track[]>(url);
    }

    async addToQueue(trackIDs: string[]) {
        const url = `${this.url}/queue`;
        return await this.fetchWithAuth<ResponseOK>(url, jsonRequestParams("POST", { ids: trackIDs }));
    }

    async updateQueue(trackIDs: string[]) {
        const url = `${this.url}/queue`;
        return await this.fetchWithAuth<ResponseOK>(url, jsonRequestParams("PUT", { ids: trackIDs }));
    }

    async removeFromQueue(trackIDs: string[]) {
        const url = `${this.url}/queue`;
        return await this.fetchWithAuth<ResponseOK>(url, jsonRequestParams("DELETE", { ids: trackIDs }));
    }

    private async fetchWithAuth<T>(url: string, params: RequestInit = {}): Promise<T> {
        params.headers = { ...params.headers, Authorization: this.authToken };

        const resp = await fetch(url, params);
        if (!resp.ok) {
            const body: ResponseErr = await resp.json();
            throw new Error(body.message);
        }

        return resp.json();
    }
}

export const airstationAPI = new AirstationAPI(API_HOST, API_PREFIX, AUTH_TOKEN);
