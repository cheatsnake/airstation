package queue

import (
	"errors"
	"testing"

	"github.com/cheatsnake/airstation/internal/track"
)

type mockStore struct {
	queueFn               func() ([]*track.Track, error)
	addToQueueFn          func(tracks []*track.Track) error
	removeFromQueueFn     func(trackIDs []string) error
	reorderQueueFn        func(trackIDs []string) error
	spinQueueFn           func() error
	currentAndNextTrackFn func() (*track.Track, *track.Track, error)
}

func (m *mockStore) Queue() ([]*track.Track, error) {
	if m.queueFn != nil {
		return m.queueFn()
	}
	return nil, nil
}

func (m *mockStore) AddToQueue(tracks []*track.Track) error {
	if m.addToQueueFn != nil {
		return m.addToQueueFn(tracks)
	}
	return nil
}

func (m *mockStore) RemoveFromQueue(trackIDs []string) error {
	if m.removeFromQueueFn != nil {
		return m.removeFromQueueFn(trackIDs)
	}
	return nil
}

func (m *mockStore) ReorderQueue(trackIDs []string) error {
	if m.reorderQueueFn != nil {
		return m.reorderQueueFn(trackIDs)
	}
	return nil
}

func (m *mockStore) SpinQueue() error {
	if m.spinQueueFn != nil {
		return m.spinQueueFn()
	}
	return nil
}

func (m *mockStore) CurrentAndNextTrack() (*track.Track, *track.Track, error) {
	if m.currentAndNextTrackFn != nil {
		return m.currentAndNextTrackFn()
	}
	return nil, nil, nil
}

func TestService_Queue(t *testing.T) {
	t.Run("returns store results", func(t *testing.T) {
		expected := []*track.Track{
			{ID: "1", Name: "Track A"},
			{ID: "2", Name: "Track B"},
		}
		mock := &mockStore{
			queueFn: func() ([]*track.Track, error) {
				return expected, nil
			},
		}
		svc := NewService(mock)
		q, err := svc.Queue()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(q) != 2 {
			t.Errorf("expected 2 tracks, got %d", len(q))
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			queueFn: func() ([]*track.Track, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewService(mock)
		_, err := svc.Queue()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_AddToQueue(t *testing.T) {
	t.Run("propagates to store", func(t *testing.T) {
		var got []*track.Track
		mock := &mockStore{
			addToQueueFn: func(tracks []*track.Track) error {
				got = tracks
				return nil
			},
		}
		svc := NewService(mock)
		input := []*track.Track{{ID: "1", Name: "Track A"}}
		err := svc.AddToQueue(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].ID != "1" {
			t.Errorf("unexpected tracks passed to store: %v", got)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			addToQueueFn: func(tracks []*track.Track) error {
				return errors.New("insert failed")
			},
		}
		svc := NewService(mock)
		err := svc.AddToQueue([]*track.Track{{ID: "1"}})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_ReorderQueue(t *testing.T) {
	t.Run("propagates to store", func(t *testing.T) {
		var got []string
		mock := &mockStore{
			reorderQueueFn: func(trackIDs []string) error {
				got = trackIDs
				return nil
			},
		}
		svc := NewService(mock)
		input := []string{"id3", "id1", "id2"}
		err := svc.ReorderQueue(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(got) != 3 || got[0] != "id3" {
			t.Errorf("unexpected IDs passed to store: %v", got)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			reorderQueueFn: func(trackIDs []string) error {
				return errors.New("reorder failed")
			},
		}
		svc := NewService(mock)
		err := svc.ReorderQueue([]string{"id1"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_RemoveFromQueue(t *testing.T) {
	t.Run("propagates to store", func(t *testing.T) {
		var got []string
		mock := &mockStore{
			removeFromQueueFn: func(trackIDs []string) error {
				got = trackIDs
				return nil
			},
		}
		svc := NewService(mock)
		input := []string{"id1", "id2"}
		err := svc.RemoveFromQueue(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(got) != 2 || got[1] != "id2" {
			t.Errorf("unexpected IDs passed to store: %v", got)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			removeFromQueueFn: func(trackIDs []string) error {
				return errors.New("remove failed")
			},
		}
		svc := NewService(mock)
		err := svc.RemoveFromQueue([]string{"id1"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_SpinQueue(t *testing.T) {
	t.Run("propagates to store", func(t *testing.T) {
		called := false
		mock := &mockStore{
			spinQueueFn: func() error {
				called = true
				return nil
			},
		}
		svc := NewService(mock)
		err := svc.SpinQueue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !called {
			t.Error("SpinQueue was not called on store")
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			spinQueueFn: func() error {
				return errors.New("spin failed")
			},
		}
		svc := NewService(mock)
		err := svc.SpinQueue()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_CurrentAndNextTrack(t *testing.T) {
	t.Run("returns store results", func(t *testing.T) {
		current := &track.Track{ID: "1", Name: "Current"}
		next := &track.Track{ID: "2", Name: "Next"}
		mock := &mockStore{
			currentAndNextTrackFn: func() (*track.Track, *track.Track, error) {
				return current, next, nil
			},
		}
		svc := NewService(mock)
		c, n, err := svc.CurrentAndNextTrack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != "1" || n.ID != "2" {
			t.Errorf("expected IDs 1/2, got %s/%s", c.ID, n.ID)
		}
	})

	t.Run("returns nil tracks when store returns nil", func(t *testing.T) {
		mock := &mockStore{
			currentAndNextTrackFn: func() (*track.Track, *track.Track, error) {
				return nil, nil, nil
			},
		}
		svc := NewService(mock)
		c, n, err := svc.CurrentAndNextTrack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c != nil || n != nil {
			t.Errorf("expected nil tracks, got %v / %v", c, n)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			currentAndNextTrackFn: func() (*track.Track, *track.Track, error) {
				return nil, nil, errors.New("db error")
			},
		}
		svc := NewService(mock)
		_, _, err := svc.CurrentAndNextTrack()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
