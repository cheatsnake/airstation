package sqlite

import (
	"testing"
)

func TestPlaylistStore_AddAndGetPlaylist(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)

	t.Run("add playlist with tracks", func(t *testing.T) {
		pl, err := inst.PlaylistStore.AddPlaylist("My Playlist", "Description", []string{a.ID, b.ID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pl.ID == "" {
			t.Error("expected non-empty ID")
		}
		if pl.Name != "My Playlist" {
			t.Errorf("expected name %q, got %q", "My Playlist", pl.Name)
		}
		if pl.TrackCount != 2 {
			t.Errorf("expected 2 tracks, got %d", pl.TrackCount)
		}
		if len(pl.Tracks) != 2 {
			t.Errorf("expected 2 track objects, got %d", len(pl.Tracks))
		}
	})

	t.Run("add playlist with no tracks", func(t *testing.T) {
		pl, err := inst.PlaylistStore.AddPlaylist("Empty Playlist", "", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pl.TrackCount != 0 {
			t.Errorf("expected 0 tracks, got %d", pl.TrackCount)
		}
	})

	t.Run("add playlist with description", func(t *testing.T) {
		pl, err := inst.PlaylistStore.AddPlaylist("Described", "A description", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pl.Description != "A description" {
			t.Errorf("expected description %q, got %q", "A description", pl.Description)
		}
	})
}

func TestPlaylistStore_Playlists(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)

	inst.PlaylistStore.AddPlaylist("Playlist 1", "", []string{a.ID})
	inst.PlaylistStore.AddPlaylist("Playlist 2", "", []string{})

	pls, err := inst.PlaylistStore.Playlists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pls) != 2 {
		t.Errorf("expected 2 playlists, got %d", len(pls))
	}
	if pls[0].TrackCount != 1 {
		t.Errorf("expected first playlist to have 1 track, got %d", pls[0].TrackCount)
	}
	if pls[1].TrackCount != 0 {
		t.Errorf("expected second playlist to have 0 tracks, got %d", pls[1].TrackCount)
	}
}

func TestPlaylistStore_GetPlaylist(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)

	t.Run("retrieve existing playlist with tracks", func(t *testing.T) {
		added, _ := inst.PlaylistStore.AddPlaylist("My Playlist", "Desc", []string{a.ID})

		pl, err := inst.PlaylistStore.Playlist(added.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pl.ID != added.ID {
			t.Errorf("expected ID %q, got %q", added.ID, pl.ID)
		}
		if pl.TrackCount != 1 || len(pl.Tracks) != 1 {
			t.Errorf("expected 1 track, got count=%d len=%d", pl.TrackCount, len(pl.Tracks))
		}
		if pl.Tracks[0].ID != a.ID {
			t.Errorf("unexpected track ID: %q", pl.Tracks[0].ID)
		}
	})

	t.Run("nonexistent playlist returns error", func(t *testing.T) {
		_, err := inst.PlaylistStore.Playlist("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent playlist, got nil")
		}
	})
}

func TestPlaylistStore_IsPlaylistExists(t *testing.T) {
	inst := setupTestDB(t)

	t.Run("returns false for missing name", func(t *testing.T) {
		exists, err := inst.PlaylistStore.IsPlaylistExists("Not Here")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected false for nonexistent playlist")
		}
	})

	t.Run("returns true after creation", func(t *testing.T) {
		inst.PlaylistStore.AddPlaylist("My Playlist", "", []string{})
		exists, err := inst.PlaylistStore.IsPlaylistExists("My Playlist")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected true for existing playlist")
		}
	})
}

func TestPlaylistStore_EditPlaylist(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)

	pl, _ := inst.PlaylistStore.AddPlaylist("Original", "Old desc", []string{a.ID})

	err := inst.PlaylistStore.EditPlaylist(pl.ID, "Updated", "New desc", []string{a.ID, b.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fetched, err := inst.PlaylistStore.Playlist(pl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "Updated" {
		t.Errorf("expected name %q, got %q", "Updated", fetched.Name)
	}
	if fetched.Description != "New desc" {
		t.Errorf("expected description %q, got %q", "New desc", fetched.Description)
	}
	if fetched.TrackCount != 2 {
		t.Errorf("expected 2 tracks, got %d", fetched.TrackCount)
	}
}

func TestPlaylistStore_DeletePlaylist(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)

	pl, _ := inst.PlaylistStore.AddPlaylist("To Delete", "", []string{a.ID})

	err := inst.PlaylistStore.DeletePlaylist(pl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = inst.PlaylistStore.Playlist(pl.ID)
	if err == nil {
		t.Error("expected error fetching deleted playlist, got nil")
	}
}
