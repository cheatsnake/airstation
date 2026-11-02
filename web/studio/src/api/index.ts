import {
    PlaybackState,
    Playlist,
    ResponseErr,
    ResponseOK,
    ScheduleRule,
    ScheduleRuleDraft,
    StationInfo,
    Track,
    TrackKind,
    TracksPage,
    VoiceOver,
    VoiceOverDraft,
} from "./types";
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

    async getTracks(
        page: number,
        limit: number,
        search: string,
        sortBy: keyof Track,
        sortOrder: "asc" | "desc",
        kind: TrackKind | "all" = "all",
    ) {
        const url = `${this.url()}/tracks?${queryParams({
            page,
            limit,
            search,
            sort_by: sortBy,
            sort_order: sortOrder,
            kind,
        })}`;
        return await this.makeRequest<TracksPage>(url);
    }

    async uploadTracks(files: File[], kind: TrackKind = "music") {
        const url = `${this.url()}/tracks`;
        const formData = new FormData();

        formData.append("kind", kind);
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

    async getCarts() {
        const url = `${this.url()}/carts`;
        return await this.makeRequest<Track[]>(url);
    }

    async playCart(id: string, mode: "next" | "now" = "next") {
        const url = `${this.url()}/carts/${id}/play`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("POST", { mode }));
    }

    async getVoiceOvers() {
        const url = `${this.url()}/voice-overs`;
        return await this.makeRequest<VoiceOver[]>(url);
    }

    async createVoiceOver(draft: VoiceOverDraft) {
        const url = `${this.url()}/voice-overs`;
        return await this.makeRequest<VoiceOver>(url, jsonRequestParams("POST", draft));
    }

    async cancelVoiceOver(id: string) {
        const url = `${this.url()}/voice-overs/${id}`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("DELETE", {}));
    }

    async getScheduleRules() {
        const url = `${this.url()}/schedule`;
        return await this.makeRequest<ScheduleRule[]>(url);
    }

    async createScheduleRule(draft: ScheduleRuleDraft) {
        const url = `${this.url()}/schedule`;
        return await this.makeRequest<ScheduleRule>(url, jsonRequestParams("POST", draft));
    }

    async editScheduleRule(id: string, draft: ScheduleRuleDraft) {
        const url = `${this.url()}/schedule/${id}`;
        return await this.makeRequest<ScheduleRule>(url, jsonRequestParams("PUT", draft));
    }

    async deleteScheduleRule(id: string) {
        const url = `${this.url()}/schedule/${id}`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("DELETE", {}));
    }

    async getLiveStatus() {
        const url = `${this.url()}/live`;
        return await this.makeRequest<{ live: boolean; mode: string; startedAt: number; segmentCount: number }>(url);
    }

    async startLiveExternal(url: string) {
        const endpoint = `${this.url()}/live/start`;
        return await this.makeRequest<ResponseOK>(endpoint, jsonRequestParams("POST", { mode: "external", url }));
    }

    async stopLive() {
        const url = `${this.url()}/live/stop`;
        return await this.makeRequest<ResponseOK>(url, jsonRequestParams("POST", {}));
    }

    /**
     * Returns the microphone ingest WebSocket URL. Browser cookies are sent with
     * the same-origin upgrade, so the JWT middleware can authorise the session.
     * WebSockets work over HTTP/1.1, unlike streaming `fetch` request bodies.
     */
    liveIngestUrl() {
        const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        return `${protocol}//${window.location.host}${this.prefix}/live/ingest`;
    }

    liveIngestSocket() {
        return new WebSocket(this.liveIngestUrl());
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

    async getStationInfo() {
        const url = `${this.url()}/station/info`;
        return await this.makeRequest<StationInfo>(url);
    }

    async editStationInfo(info: StationInfo) {
        const url = `${this.url()}/station/info`;
        return await this.makeRequest<StationInfo>(url, jsonRequestParams("PUT", info));
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
