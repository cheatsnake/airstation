package queue

import "github.com/cheatsnake/airstation/internal/track"

type Store interface {
	Queue() ([]*track.Track, error)
	AddToQueue(tracks []*track.Track) error
	RemoveFromQueue(trackIDs []string) error
	ReorderQueue(trackIDs []string) error
	SpinQueue() error
	CurrentAndNextTrack() (*track.Track, *track.Track, error)
}
