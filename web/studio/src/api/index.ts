import { PlaybackState, Playlist, ResponseErr, ResponseOK, Track, TracksPage } from "./types";
import { jsonRequestParams, queryParams } from "./utils";

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

    async login(secret: string) {
        const url = `${this.url()}/login`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("POST", { secret }));
    }

    async getPlayback() {
        const url = `${this.url()}/playback`;
        return await this.makeRequest<PlaybackState>(url);
    }

    async pausePlayback() {
        const url = `${this.url()}/playback/pause`;
        return await this.makeRequest<PlaybackState>(url, jsonRequestParams("POST", {}));
    }

    async playPlayback() {
        const url = `${this.url()}/playback/play`;
        return await this.makeRequest<PlaybackState>(url, jsonRequestParams("POST", {}));
    }

    async getTracks(page: number, limit: number, search: string, sortBy: keyof Track, sortOrder: "asc" | "desc") {
        const url = `${this.url()}/tracks?${queryParams({
            page,
            limit,
            search,
            sort_by: sortBy,
            sort_order: sortOrder,
        })}`;
        return await this.makeRequest<TracksPage>(url);
    }

    async uploadTracks(files: File[]) {
        const url = `${this.url()}/tracks`;
        const formData = new FormData();

        for (let i = 0; i < files.length; i++) {
            formData.append("tracks", files[i]);
        }

        return await this.makeRequest<ResponseOK>(url, {
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
        const url = `${this.url()}/queue`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("POST", { ids: trackIDs }));
    }

    async updateQueue(trackIDs: string[]) {
        const url = `${this.url()}/queue`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("PUT", { ids: trackIDs }));
    }

    async removeFromQueue(trackIDs: string[]) {
        const url = `${this.url()}/queue`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("DELETE", { ids: trackIDs }));
    }

    async addPlaylist(name: string, trackIDs: string[], description?: string) {
        const url = `${this.url()}/playlist`;
        return await this.makeRequest<Playlist>(url, jsonRequestParams("POST", { name, description, trackIDs }));
    }

    async getPlaylists() {
        const url = `${this.url()}/playlists`;
        return await this.makeRequest<Playlist[]>(url);
    }

    async getPlaylist(id: string) {
        const url = `${this.url()}/playlist/` + id;
        return await this.makeRequest<Playlist>(url);
    }

    async editPlaylist(id: string, name: string, trackIDs: string[], description?: string) {
        const url = `${this.url()}/playlist/` + id;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("PUT", { name, description, trackIDs }));
    }

    async deletePlaylist(id: string) {
        const url = `${this.url()}/playlist/` + id;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("DELETE", {}));
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
