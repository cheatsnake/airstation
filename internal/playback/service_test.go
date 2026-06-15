package playback

import (
	"errors"
	"testing"
)

type mockStore struct {
	addPlaybackHistoryFn      func(playedAt int64, trackName string) error
	recentPlaybackHistoryFn   func(limit int) ([]*History, error)
	deleteOldPlaybackHistoryFn func() (int64, error)
}

func (m *mockStore) AddPlaybackHistory(playedAt int64, trackName string) error {
	if m.addPlaybackHistoryFn != nil {
		return m.addPlaybackHistoryFn(playedAt, trackName)
	}
	return nil
}

func (m *mockStore) RecentPlaybackHistory(limit int) ([]*History, error) {
	if m.recentPlaybackHistoryFn != nil {
		return m.recentPlaybackHistoryFn(limit)
	}
	return nil, nil
}

func (m *mockStore) DeleteOldPlaybackHistory() (int64, error) {
	if m.deleteOldPlaybackHistoryFn != nil {
		return m.deleteOldPlaybackHistoryFn()
	}
	return 0, nil
}

func TestService_AddPlaybackHistory(t *testing.T) {
	t.Run("calls store with correct track name", func(t *testing.T) {
		var gotName string
		mock := &mockStore{
			addPlaybackHistoryFn: func(playedAt int64, trackName string) error {
				gotName = trackName
				return nil
			},
		}
		svc := NewService(mock)
		svc.AddPlaybackHistory("Test Track")
		if gotName != "Test Track" {
			t.Errorf("expected track name %q, got %q", "Test Track", gotName)
		}
	})
}

func TestService_RecentPlaybackHistory(t *testing.T) {
	t.Run("returns store results", func(t *testing.T) {
		expected := []*History{
			{ID: 1, PlayedAt: 1000, TrackName: "Track A"},
			{ID: 2, PlayedAt: 2000, TrackName: "Track B"},
		}
		mock := &mockStore{
			recentPlaybackHistoryFn: func(limit int) ([]*History, error) {
				return expected, nil
			},
		}
		svc := NewService(mock)
		history, err := svc.RecentPlaybackHistory(50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(history) != 2 {
			t.Errorf("expected 2 history entries, got %d", len(history))
		}
		if history[0].TrackName != "Track A" {
			t.Errorf("expected Track A, got %s", history[0].TrackName)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			recentPlaybackHistoryFn: func(limit int) ([]*History, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewService(mock)
		_, err := svc.RecentPlaybackHistory(50)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_DeleteOldPlaybackHistory(t *testing.T) {
	t.Run("calls store", func(t *testing.T) {
		called := false
		mock := &mockStore{
			deleteOldPlaybackHistoryFn: func() (int64, error) {
				called = true
				return 5, nil
			},
		}
		svc := NewService(mock)
		svc.DeleteOldPlaybackHistory()
		if !called {
			t.Error("DeleteOldPlaybackHistory was not called on store")
		}
	})
}
