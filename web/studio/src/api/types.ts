export interface Track {
    id: string;
    name: string;
    path: string;
    duration: number;
    bitRate: number;
}

export interface TracksPage {
    tracks: Track[];
    page: number;
    limit: number;
    total: number;
}

export interface PlaybackState {
    currentTrack: Track | null;
    currentTrackElapsed: number;
    isPlaying: boolean;
    updatedAt: number;
}

export interface ResponseErr {
    message: string;
}

export interface ResponseOK {
    message: string;
}

export interface Playlist {
    id: string;
    name: string;
    description?: string;
    tracks: Track[];
    trackCount: number;
}
