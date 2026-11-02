export interface Track {
    id: string;
    name: string;
    path: string;
    duration: number;
    bitRate: number;
    kind?: string;
}

export interface VoiceOver {
    id: string;
    voiceName?: string;
    target: "next" | "current";
    status: string;
    duckDb: number;
}

export interface PlaybackState {
    currentTrack: Track | null;
    currentTrackElapsed: number;
    currentVoiceOver?: VoiceOver | null;
    nextTrack?: Track | null;
    isPlaying: boolean;
    mode?: string;
}

export interface PlaybackHistory {
    id: number;
    playedAt: number;
    trackName: string;
}

export interface StationInfo {
    name: string;
    description: string;
    faviconURL: string;
    logoURL: string;
    location: string;
    timezone: string;
    links: string;
    theme: string;
    audioBitRate?: number;
}

export interface ResponseErr {
    message: string;
}

export interface ResponseOK {
    message: string;
}
