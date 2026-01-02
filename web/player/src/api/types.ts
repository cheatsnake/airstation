export interface Track {
    id: string;
    name: string;
    path: string;
    duration: number;
    bitRate: number;
}

export interface PlaybackState {
    currentTrack: Track | null;
    currentTrackElapsed: number;
    isPlaying: boolean;
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
}

export interface ResponseErr {
    message: string;
}

export interface ResponseOK {
    message: string;
}
