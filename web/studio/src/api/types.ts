export type TrackKind = "music" | "jingle" | "ad" | "station_id" | "voice";

export const TRACK_KINDS: { value: TrackKind; label: string; color: string }[] = [
    { value: "music", label: "Music", color: "air" },
    { value: "jingle", label: "Jingle", color: "teal" },
    { value: "ad", label: "Ad", color: "yellow" },
    { value: "station_id", label: "Station ID", color: "violet" },
    { value: "voice", label: "Voice link", color: "green" },
];

export const trackKindLabel = (kind?: string) => TRACK_KINDS.find((k) => k.value === kind)?.label ?? "Music";
export const trackKindColor = (kind?: string) => TRACK_KINDS.find((k) => k.value === kind)?.color ?? "air";

export interface Track {
    id: string;
    name: string;
    path: string;
    duration: number;
    bitRate: number;
    kind: TrackKind;
    fadeIn?: number;
    fadeOut?: number;
    duckDb?: number;
}

export interface TracksPage {
    tracks: Track[];
    page: number;
    limit: number;
    total: number;
}

export type PlaybackMode = "automation" | "live" | "relay";

export interface VoiceOver {
    id: string;
    voiceTrackId: string;
    voiceName?: string;
    target: "next" | "current";
    targetTrackId?: string;
    fadeIn: number;
    fadeOut: number;
    duckDb: number;
    startOffset: number;
    status: "pending" | "rendered" | "played" | "failed" | "cancelled";
    appliedTrackId?: string;
    createdAt: number;
}

export interface VoiceOverDraft {
    voiceTrackId: string;
    target: "next" | "current";
    fadeIn: number;
    fadeOut: number;
    duckDb: number;
    startOffset: number;
}

export interface PlaybackState {
    currentTrack: Track | null;
    currentTrackElapsed: number;
    currentVoiceOver?: VoiceOver | null;
    nextTrack?: Track | null;
    isPlaying: boolean;
    mode: PlaybackMode;
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

export interface ScheduleRule {
    id: string;
    name: string;
    kind: TrackKind;
    everyNTracks: number;
    atMinute: number;
    enabled: boolean;
    createdAt: number;
}

export interface ScheduleRuleDraft {
    name: string;
    kind: TrackKind;
    everyNTracks: number;
    atMinute: number;
    enabled: boolean;
}
