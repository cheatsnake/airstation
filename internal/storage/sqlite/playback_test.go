package sqlite

import (
	"testing"
	"time"
)

func TestPlaybackStore_AddPlaybackHistory(t *testing.T) {
	inst := setupTestDB(t)

	now := time.Now().Unix()
	err := inst.PlaybackStore.AddPlaybackHistory(now, "Test Track")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlaybackStore_RecentPlaybackHistory(t *testing.T) {
	inst := setupTestDB(t)

	now := time.Now().Unix()
	inst.PlaybackStore.AddPlaybackHistory(now-100, "Track A")
	inst.PlaybackStore.AddPlaybackHistory(now-50, "Track B")
	inst.PlaybackStore.AddPlaybackHistory(now, "Track C")

	t.Run("returns limited results", func(t *testing.T) {
		history, err := inst.PlaybackStore.RecentPlaybackHistory(2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(history) != 2 {
			t.Errorf("expected 2 entries, got %d", len(history))
		}
	})

	t.Run("returns in descending order", func(t *testing.T) {
		history, err := inst.PlaybackStore.RecentPlaybackHistory(10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(history) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(history))
		}
		if history[0].TrackName != "Track C" {
			t.Errorf("expected most recent first, got %q", history[0].TrackName)
		}
		if history[2].TrackName != "Track A" {
			t.Errorf("expected oldest last, got %q", history[2].TrackName)
		}
	})
}

func TestPlaybackStore_DeleteOldPlaybackHistory(t *testing.T) {
	inst := setupTestDB(t)

	now := time.Now().Unix()
	thirtyOneDaysAgo := now - 31*24*60*60
	recentTime := now - 100

	inst.PlaybackStore.AddPlaybackHistory(thirtyOneDaysAgo, "Old Track")
	inst.PlaybackStore.AddPlaybackHistory(recentTime, "Recent Track")

	deleted, err := inst.PlaybackStore.DeleteOldPlaybackHistory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	history, err := inst.PlaybackStore.RecentPlaybackHistory(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 remaining entry, got %d", len(history))
	}
	if history[0].TrackName != "Recent Track" {
		t.Errorf("expected Recent Track remaining, got %q", history[0].TrackName)
	}
}
