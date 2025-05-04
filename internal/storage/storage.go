package storage

import "github.com/cheatsnake/airstation/internal/track"

type Storage interface {
	TrackStore
}

type TrackStore interface {
	Tracks(page, limit int, search, sortBy, sortOrder string) ([]*track.Track, int, error)
	TrackByID(ID string) (*track.Track, error)
	TracksByIDs(IDs []string) ([]*track.Track, error)
	AddTrack(name, path string, duration float64, bitRate int) (*track.Track, error)
	DeleteTracks(IDs []string) error
	EditTrack(track *track.Track) (*track.Track, error)

	Queue() ([]*track.Track, error)
	AddToQueue(tracks []*track.Track) error
	RemoveFromQueue(trackIDs []string) error
	ReorderQueue(trackIDs []string) error
	SpinQueue() error
	CurrentAndNextTrack() (*track.Track, *track.Track, error)

	AddPlaybackHistory(playedAt int64, trackName string) error
	RecentPlaybackHistory(limit int) ([]*track.PlaybackHistory, error)
	DeleteOldPlaybackHistory() (int64, error)

	Close() error
}
