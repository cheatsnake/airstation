package track

import (
	"fmt"
	"slices"
	"sync"
)

type Queue struct {
	Tracks []Track
	mutex  sync.Mutex
}

func NewQueue(tracks []Track) *Queue {
	return &Queue{
		Tracks: tracks,
	}
}

func (q *Queue) Add(tracks []Track) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.Tracks = append(q.Tracks, tracks...)
}

func (q *Queue) Remove(trackIDs []string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	filtered := make([]Track, 0, len(q.Tracks)-len(trackIDs))

	for _, t := range q.Tracks {
		if !slices.Contains(trackIDs, t.ID) {
			filtered = append(filtered, t)
		}
	}

	q.Tracks = filtered
}

func (q *Queue) Reorder(trackIDs []string) error {
	ordered := make([]Track, 0, len(trackIDs))

	for _, id := range trackIDs {
		track := q.FindTrack(id)
		if track == nil {
			return fmt.Errorf("track with ID %s does not exist in the queue", id)
		}
		ordered = append(ordered, *track)
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Tracks = ordered

	return nil
}

func (q *Queue) FindTrack(trackID string) *Track {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for _, t := range q.Tracks {
		if t.ID == trackID {
			return &t
		}
	}

	return nil
}

func (q *Queue) CurrentTrack() *Track {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.Tracks) == 0 {
		return nil
	}

	return &q.Tracks[0]
}

func (q *Queue) NextTrack() *Track {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.Tracks) == 0 {
		return nil
	}

	if len(q.Tracks) == 1 {
		return &q.Tracks[0]
	}

	return &q.Tracks[1]
}

func (q *Queue) Spin() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.Tracks = append(q.Tracks[1:], q.Tracks[0])
}
