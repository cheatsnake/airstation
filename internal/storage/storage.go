package storage

import "github.com/cheatsnake/airstation/internal/track"

type Storage interface {
	TrackStore
}

type TrackStore interface {
	GetTracks(page, limit int) ([]*track.Track, int, error)
	FindTrack(ID string) (*track.Track, error)
	FindTracks(IDs []string) ([]*track.Track, error)
	AddTrack(name, path string, duration float64, bitRate int) (*track.Track, error)
	RemoveTracks(IDs []string) error
	EditTrack(track *track.Track) (*track.Track, error)

	GetQueue() ([]*track.Track, error)
	AddToQueue(tracks []*track.Track) error
	DeleteFromQueue(trackIDs []string) error
	ReorderQueue(trackIDs []string) error
	SpinQueue() error
	CurrentAndNextTrack() (*track.Track, *track.Track, error)

	Close() error
}
