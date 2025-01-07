package track

import (
	"testing"
)

func TestQueue(t *testing.T) {
	t.Run("NewQueue initializes queue with tracks", func(t *testing.T) {
		tracks := []Track{
			{ID: "1", Name: "Track 1"},
			{ID: "2", Name: "Track 2"},
		}

		q := NewQueue(tracks)

		if len(q.Tracks) != len(tracks) {
			t.Errorf("expected %d tracks, got %d", len(tracks), len(q.Tracks))
		}
	})

	t.Run("Add appends new tracks to the queue", func(t *testing.T) {
		q := NewQueue([]Track{})

		q.Add([]Track{
			{ID: "1", Name: "Track 1"},
			{ID: "2", Name: "Track 2"},
		})

		if len(q.Tracks) != 2 {
			t.Errorf("expected 2 tracks, got %d", len(q.Tracks))
		}
	})

	t.Run("Remove deletes tracks by IDs", func(t *testing.T) {
		q := NewQueue([]Track{
			{ID: "1", Name: "Track 1"},
			{ID: "2", Name: "Track 2"},
			{ID: "3", Name: "Track 3"},
		})

		q.Remove([]string{"2"})

		if len(q.Tracks) != 2 {
			t.Errorf("expected 2 tracks, got %d", len(q.Tracks))
		}

		if q.Tracks[0].ID != "1" || q.Tracks[1].ID != "3" {
			t.Errorf("unexpected remaining tracks: %+v", q.Tracks)
		}
	})

	t.Run("Reorder rearranges tracks by IDs", func(t *testing.T) {
		q := NewQueue([]Track{
			{ID: "1", Name: "Track 1"},
			{ID: "2", Name: "Track 2"},
			{ID: "3", Name: "Track 3"},
		})

		newOrder := []string{"3", "1", "2"}

		err := q.Reorder(newOrder)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i, track := range q.Tracks {
			if track.ID != newOrder[i] {
				t.Errorf("unexpected order, got: %+v", q.Tracks)
			}
		}
	})

	t.Run("Reorder returns error if a track ID does not exist", func(t *testing.T) {
		q := NewQueue([]Track{
			{ID: "1", Name: "Track 1"},
			{ID: "2", Name: "Track 2"},
		})

		err := q.Reorder([]string{"3", "1"})
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("FindTrack returns the correct track by ID", func(t *testing.T) {
		q := NewQueue([]Track{
			{ID: "1", Name: "Track 1"},
			{ID: "2", Name: "Track 2"},
		})

		track := q.FindTrack("2")

		if track == nil || track.ID != "2" {
			t.Errorf("expected track with ID 2, got: %+v", track)
		}
	})

	t.Run("CurrentTrack returns the first track in the queue", func(t *testing.T) {
		q := NewQueue([]Track{
			{ID: "1", Name: "Track 1"},
			{ID: "2", Name: "Track 2"},
		})

		track := q.CurrentTrack()

		if track == nil || track.ID != "1" {
			t.Errorf("expected track with ID 1, got: %+v", track)
		}
	})

	t.Run("NextTrack returns the second track in the queue", func(t *testing.T) {
		q := NewQueue([]Track{
			{ID: "1", Name: "Track 1"},
			{ID: "2", Name: "Track 2"},
		})

		track := q.NextTrack()

		if track == nil || track.ID != "2" {
			t.Errorf("expected track with ID 2, got: %+v", track)
		}
	})

	t.Run("NextTrack returns the first track if only one exists", func(t *testing.T) {
		q := NewQueue([]Track{
			{ID: "1", Name: "Track 1"},
		})

		track := q.NextTrack()

		if track == nil || track.ID != "1" {
			t.Errorf("expected track with ID 1, got: %+v", track)
		}
	})

	t.Run("Spin moves the first track to the end of the queue", func(t *testing.T) {
		q := NewQueue([]Track{
			{ID: "1", Name: "Track 1"},
			{ID: "2", Name: "Track 2"},
			{ID: "3", Name: "Track 3"},
		})

		q.Spin()

		expectedOrder := []string{"2", "3", "1"}

		for i, track := range q.Tracks {
			if track.ID != expectedOrder[i] {
				t.Errorf("unexpected order after spin, got: %+v", q.Tracks)
			}
		}
	})

	t.Run("Spin do nothing if only one track exists", func(t *testing.T) {
		q := NewQueue([]Track{
			{ID: "1", Name: "Track 1"},
		})

		q.Spin()

		expectedOrder := []string{"1"}

		for i, track := range q.Tracks {
			if track.ID != expectedOrder[i] {
				t.Errorf("unexpected order after spin, got: %+v", q.Tracks)
			}
		}
	})

	t.Run("Spin do nothing if queue is empty", func(t *testing.T) {
		q := NewQueue([]Track{})

		q.Spin()

		if len(q.Tracks) != 0 {
			t.Errorf("unexpected queue length, expected 0, got: %d", len(q.Tracks))
		}
	})
}
