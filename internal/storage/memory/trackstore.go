package memory

import (
	"fmt"
	"slices"
	"sync"

	"github.com/cheatsnake/airstation/internal/track"
)

type TrackStore struct {
	tracks []track.Track
	mutex  sync.Mutex
}

func NewTrackStore() *TrackStore {
	return &TrackStore{
		tracks: make([]track.Track, 10),
	}
}

func (ts *TrackStore) AddTrack(name, path string, duration float64, bitRate int) (*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	t := track.New(name, path, duration, bitRate)
	ts.tracks = append(ts.tracks, *t)

	return t, nil
}

func (ts *TrackStore) RemoveTracks(IDs []string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	filtered := make([]track.Track, 0, len(ts.tracks)-len(IDs))

	for _, t := range ts.tracks {
		if slices.Contains(IDs, t.ID) {
			continue
		}

		filtered = append(filtered, t)
	}

	ts.tracks = filtered

	return nil
}

func (ts *TrackStore) EditTrack(track track.Track) (*track.Track, error) {
	t, err := ts.FindTrack(track.ID)
	if err != nil {
		return nil, err
	}

	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if track.Name != "" {
		t.Name = track.Name
	}

	if track.Path != "" {
		t.Path = track.Path
	}

	if track.Duration != 0 {
		t.Duration = track.Duration
	}

	if track.BitRate != 0 {
		t.BitRate = track.BitRate
	}

	return t, nil
}

func (ts *TrackStore) FindTrack(ID string) (*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	for _, t := range ts.tracks {
		if t.ID == ID {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("track with ID %s not found", ID)
}

func (ts *TrackStore) FindTracks(IDs []string) ([]*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	found := make([]*track.Track, 0, len(IDs))

	for _, t := range ts.tracks {
		if slices.Contains(IDs, t.ID) {
			found = append(found, &t)
		}
	}

	return found, nil

}
