package track

import (
	"fmt"
	"slices"
)

// Queue represents a collection of audio tracks.
type Queue struct {
	Tracks []Track `json:"tracks"` // A slice containing the tracks currently in the queue.
}

// NewQueue creates and returns a new Queue instance initialized with a given list of tracks.
//
// Parameters:
//   - tracks: A slice of Track instances to initialize the queue.
//
// Returns:
//   - A pointer to a newly created Queue instance.
func NewQueue(tracks []Track) *Queue {
	return &Queue{
		Tracks: tracks,
	}
}

// Add appends new tracks to the end of the queue.
//
// Parameters:
//   - tracks: A slice of Track instances to add to the queue.
func (q *Queue) Add(tracks []Track) {
	q.Tracks = append(q.Tracks, tracks...)
}

// Remove removes tracks from the queue based on their track IDs.
//
// Parameters:
//   - trackIDs: A slice of track IDs to remove from the queue.
func (q *Queue) Remove(trackIDs []string) {
	filtered := make([]Track, 0, len(q.Tracks)-len(trackIDs))

	for _, t := range q.Tracks {
		if !slices.Contains(trackIDs, t.ID) {
			filtered = append(filtered, t)
		}
	}

	q.Tracks = filtered
}

// Reorder reorders the tracks in the queue based on the provided list of track IDs.
//
// Parameters:
//   - trackIDs: A slice of track IDs representing the desired order of tracks in the queue.
//
// Returns:
//   - An error if any track ID does not exist in the queue.
func (q *Queue) Reorder(trackIDs []string) error {
	ordered := make([]Track, 0, len(trackIDs))

	for _, id := range trackIDs {
		track := q.FindTrack(id)
		if track == nil {
			return fmt.Errorf("track with ID %s does not exist in the queue", id)
		}
		ordered = append(ordered, *track)
	}

	q.Tracks = ordered

	return nil
}

// FindTrack finds and returns the track with the given track ID, or nil if not found.
//
// Parameters:
//   - trackID: The ID of the track to search for.
//
// Returns:
//   - A pointer to the found Track, or nil if no matching track is found.
func (q *Queue) FindTrack(trackID string) *Track {
	for _, t := range q.Tracks {
		if t.ID == trackID {
			return &t
		}
	}

	return nil
}

// CurrentTrack returns the first track in the queue, or nil if the queue is empty.
//
// Returns:
//   - A pointer to the first track in the queue, or nil if the queue is empty.
func (q *Queue) CurrentTrack() *Track {
	if len(q.Tracks) == 0 {
		return nil
	}

	return &q.Tracks[0]
}

// NextTrack returns the second track in the queue, or the first if there is only one track, or nil if the queue is empty.
//
// Returns:
//   - A pointer to the second track, or the first track if there's only one, or nil if the queue is empty.
func (q *Queue) NextTrack() *Track {
	if len(q.Tracks) == 0 {
		return nil
	}

	if len(q.Tracks) == 1 {
		return &q.Tracks[0]
	}

	return &q.Tracks[1]
}

// Spin shifts the queue by moving the first track to the end of the queue, effectively rotating the queue by one position.
func (q *Queue) Spin() {
	if len(q.Tracks) < 2 {
		return
	}

	q.Tracks = append(q.Tracks[1:], q.Tracks[0])
}
