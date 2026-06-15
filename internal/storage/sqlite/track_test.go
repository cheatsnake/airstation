package sqlite

import (
	"io"
	"log/slog"
	"testing"

	"github.com/cheatsnake/airstation/internal/track"
)

func setupTestDB(t *testing.T) *Instance {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	inst, err := New(":memory:", log)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() { inst.Close() })
	return inst
}

func addTestTrack(t *testing.T, inst *Instance, name, path string, duration float64, bitRate int) *track.Track {
	t.Helper()
	tr, err := inst.TrackStore.AddTrack(name, path, duration, bitRate)
	if err != nil {
		t.Fatalf("failed to add test track: %v", err)
	}
	return tr
}

func TestTrackStore_AddTrack(t *testing.T) {
	inst := setupTestDB(t)

	tr, err := inst.TrackStore.AddTrack("Test Track", "/tracks/test.aac", 120.0, 192)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.ID == "" {
		t.Error("expected non-empty ID")
	}
	if tr.Name != "Test Track" {
		t.Errorf("expected name %q, got %q", "Test Track", tr.Name)
	}
	if tr.Path != "/tracks/test.aac" {
		t.Errorf("expected path %q, got %q", "/tracks/test.aac", tr.Path)
	}
	if tr.Duration != 120.0 {
		t.Errorf("expected duration %f, got %f", 120.0, tr.Duration)
	}
	if tr.BitRate != 192 {
		t.Errorf("expected bitrate %d, got %d", 192, tr.BitRate)
	}
}

func TestTrackStore_TrackByID(t *testing.T) {
	inst := setupTestDB(t)
	added := addTestTrack(t, inst, "Track A", "/tracks/a.aac", 60.0, 128)

	t.Run("retrieve existing track", func(t *testing.T) {
		tr, err := inst.TrackStore.TrackByID(added.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tr.ID != added.ID {
			t.Errorf("expected ID %q, got %q", added.ID, tr.ID)
		}
		if tr.Name != "Track A" {
			t.Errorf("expected name %q, got %q", "Track A", tr.Name)
		}
	})

	t.Run("track not found returns error", func(t *testing.T) {
		_, err := inst.TrackStore.TrackByID("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent track, got nil")
		}
	})
}

func TestTrackStore_Tracks(t *testing.T) {
	inst := setupTestDB(t)
	addTestTrack(t, inst, "Alpha Track", "/tracks/a.aac", 60.0, 128)
	addTestTrack(t, inst, "Beta Song", "/tracks/b.aac", 180.0, 192)
	addTestTrack(t, inst, "Charlie Chant", "/tracks/c.aac", 90.0, 256)

	t.Run("paginate with defaults", func(t *testing.T) {
		tracks, total, err := inst.TrackStore.Tracks(1, 20, "", "id", "asc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(tracks) != 3 {
			t.Errorf("expected 3 tracks, got %d", len(tracks))
		}
	})

	t.Run("pagination with limit and offset", func(t *testing.T) {
		tracks, total, err := inst.TrackStore.Tracks(2, 1, "", "id", "asc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
		if len(tracks) != 1 {
			t.Errorf("expected 1 track on page 2, got %d", len(tracks))
		}
	})

	t.Run("search by name", func(t *testing.T) {
		tracks, total, err := inst.TrackStore.Tracks(1, 20, "alpha", "id", "asc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1 for search 'alpha', got %d", total)
		}
		if len(tracks) != 1 || tracks[0].Name != "Alpha Track" {
			t.Errorf("expected Alpha Track, got %v", tracks)
		}
	})

	t.Run("sort by name desc", func(t *testing.T) {
		tracks, _, err := inst.TrackStore.Tracks(1, 20, "", "name", "desc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tracks[0].Name != "Charlie Chant" {
			t.Errorf("expected Charlie Chant first with desc sort, got %q", tracks[0].Name)
		}
	})

	t.Run("sort by duration asc", func(t *testing.T) {
		tracks, _, err := inst.TrackStore.Tracks(1, 20, "", "duration", "asc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tracks[0].Duration != 60.0 {
			t.Errorf("expected shortest first, got duration %f", tracks[0].Duration)
		}
	})
}

func TestTrackStore_TracksByIDs(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)
	addTestTrack(t, inst, "Track C", "/c.aac", 180.0, 256)

	tracks, err := inst.TrackStore.TracksByIDs([]string{a.ID, b.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 2 {
		t.Errorf("expected 2 tracks, got %d", len(tracks))
	}
}

func TestTrackStore_EditTrack(t *testing.T) {
	inst := setupTestDB(t)
	tr := addTestTrack(t, inst, "Original", "/tracks/orig.aac", 60.0, 128)

	tr.Name = "Updated"
	tr.Path = "/tracks/updated.aac"
	tr.Duration = 90.0
	tr.BitRate = 256

	edited, err := inst.TrackStore.EditTrack(tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if edited.Name != "Updated" {
		t.Errorf("expected name %q, got %q", "Updated", edited.Name)
	}

	fetched, err := inst.TrackStore.TrackByID(tr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "Updated" {
		t.Errorf("persisted name %q, want %q", fetched.Name, "Updated")
	}
}

func TestTrackStore_DeleteTracks(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)

	err := inst.TrackStore.DeleteTracks([]string{a.ID, b.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = inst.TrackStore.TrackByID(a.ID)
	if err == nil {
		t.Error("expected error fetching deleted track, got nil")
	}
}
