package storage

import "github.com/cheatsnake/airstation/internal/track"

type TrackStore interface {
	AddTrack(name, path string, duration float64, bitrate int) (*track.Track, error)

	RemoveTracks(IDs []string) error

	EditTrack(track track.Track) (*track.Track, error)

	FindTrack(ID string) (*track.Track, error)
	FindTracks(IDs []string) ([]*track.Track, error)
}
