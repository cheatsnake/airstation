import { PlaybackState, ResponseErr, Track, TracksPage } from "./types";
import { queryParams } from "./utils";

const API_HOST = "";
const API_PREFIX = "/v1/api";
const AUTH_TOKEN = "test111111";

class AirstationAPI {
  private host: string;
  private prefix: string;
  private authToken: string;

  constructor(host: string, prefix: string, authToken: string) {
    this.host = host;
    this.prefix = prefix;
    this.authToken = authToken;
  }

  async getPlayback() {
    const url = `${this.host + this.prefix}/playback`;
    return await this.fetchWithAuth<PlaybackState>(url);
  }

  async getTracks(page: number, limit: number, search: string) {
    const url = `${this.host + this.prefix}/tracks?${queryParams({ page, limit, search })}`;
    return await this.fetchWithAuth<TracksPage>(url);
  }

  async getQueue() {
    const url = `${this.host + this.prefix}/queue`;
    return await this.fetchWithAuth<Track[]>(url);
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
