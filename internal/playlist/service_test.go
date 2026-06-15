package playlist

import (
	"errors"
	"testing"
)

type mockStore struct {
	addPlaylistFn      func(name, description string, trackIDs []string) (*Playlist, error)
	playlistsFn        func() ([]*Playlist, error)
	playlistFn         func(id string) (*Playlist, error)
	isPlaylistExistsFn func(name string) (bool, error)
	editPlaylistFn     func(id, name, description string, trackIDs []string) error
	deletePlaylistFn   func(id string) error
}

func (m *mockStore) AddPlaylist(name, description string, trackIDs []string) (*Playlist, error) {
	if m.addPlaylistFn != nil {
		return m.addPlaylistFn(name, description, trackIDs)
	}
	return &Playlist{ID: "pl-1", Name: name, Description: description}, nil
}

func (m *mockStore) Playlists() ([]*Playlist, error) {
	if m.playlistsFn != nil {
		return m.playlistsFn()
	}
	return nil, nil
}

func (m *mockStore) Playlist(id string) (*Playlist, error) {
	if m.playlistFn != nil {
		return m.playlistFn(id)
	}
	return nil, nil
}

func (m *mockStore) IsPlaylistExists(name string) (bool, error) {
	if m.isPlaylistExistsFn != nil {
		return m.isPlaylistExistsFn(name)
	}
	return false, nil
}

func (m *mockStore) EditPlaylist(id, name, description string, trackIDs []string) error {
	if m.editPlaylistFn != nil {
		return m.editPlaylistFn(id, name, description, trackIDs)
	}
	return nil
}

func (m *mockStore) DeletePlaylist(id string) error {
	if m.deletePlaylistFn != nil {
		return m.deletePlaylistFn(id)
	}
	return nil
}

func TestService_AddPlaylist(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		svc := NewService(&mockStore{})
		pl, err := svc.AddPlaylist("My Playlist", "A description", []string{"id1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pl.Name != "My Playlist" {
			t.Errorf("expected name %q, got %q", "My Playlist", pl.Name)
		}
	})

	t.Run("rejects name too short", func(t *testing.T) {
		svc := NewService(&mockStore{})
		_, err := svc.AddPlaylist("ab", "", []string{"id1"})
		if err == nil {
			t.Error("expected error for short name, got nil")
		}
	})

	t.Run("rejects description too long", func(t *testing.T) {
		svc := NewService(&mockStore{})
		_, err := svc.AddPlaylist("My Playlist", string(make([]byte, maxDescrLen+1)), []string{"id1"})
		if err == nil {
			t.Error("expected error for long description, got nil")
		}
	})

	t.Run("rejects empty track ID", func(t *testing.T) {
		svc := NewService(&mockStore{})
		_, err := svc.AddPlaylist("My Playlist", "", []string{""})
		if err == nil {
			t.Error("expected error for empty track ID, got nil")
		}
	})

	t.Run("rejects too many tracks", func(t *testing.T) {
		svc := NewService(&mockStore{})
		ids := make([]string, maxTracks+1)
		for i := range ids {
			ids[i] = "id"
		}
		_, err := svc.AddPlaylist("My Playlist", "", ids)
		if err == nil {
			t.Error("expected error for too many tracks, got nil")
		}
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		mock := &mockStore{
			isPlaylistExistsFn: func(name string) (bool, error) {
				return true, nil
			},
		}
		svc := NewService(mock)
		_, err := svc.AddPlaylist("Existing Name", "", []string{"id1"})
		if err == nil {
			t.Error("expected error for duplicate name, got nil")
		}
	})

	t.Run("propagates store IsPlaylistExists error", func(t *testing.T) {
		mock := &mockStore{
			isPlaylistExistsFn: func(name string) (bool, error) {
				return false, errors.New("db error")
			},
		}
		svc := NewService(mock)
		_, err := svc.AddPlaylist("My Playlist", "", []string{"id1"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("propagates store AddPlaylist error", func(t *testing.T) {
		mock := &mockStore{
			addPlaylistFn: func(name, description string, trackIDs []string) (*Playlist, error) {
				return nil, errors.New("insert failed")
			},
		}
		svc := NewService(mock)
		_, err := svc.AddPlaylist("My Playlist", "", []string{"id1"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_Playlists(t *testing.T) {
	t.Run("returns store results", func(t *testing.T) {
		expected := []*Playlist{
			{ID: "1", Name: "Pl1"},
			{ID: "2", Name: "Pl2"},
		}
		mock := &mockStore{
			playlistsFn: func() ([]*Playlist, error) {
				return expected, nil
			},
		}
		svc := NewService(mock)
		pls, err := svc.Playlists()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pls) != 2 {
			t.Errorf("expected 2 playlists, got %d", len(pls))
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			playlistsFn: func() ([]*Playlist, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewService(mock)
		_, err := svc.Playlists()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_Playlist(t *testing.T) {
	expected := &Playlist{ID: "1", Name: "Pl1"}

	t.Run("returns store result", func(t *testing.T) {
		mock := &mockStore{
			playlistFn: func(id string) (*Playlist, error) {
				return expected, nil
			},
		}
		svc := NewService(mock)
		pl, err := svc.Playlist("1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pl.ID != "1" {
			t.Errorf("expected ID %q, got %q", "1", pl.ID)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			playlistFn: func(id string) (*Playlist, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewService(mock)
		_, err := svc.Playlist("missing")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_EditPlaylist(t *testing.T) {
	t.Run("successful edit", func(t *testing.T) {
		mock := &mockStore{
			editPlaylistFn: func(id, name, description string, trackIDs []string) error {
				return nil
			},
		}
		svc := NewService(mock)
		err := svc.EditPlaylist("1", "New Name", "New desc", []string{"id1"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("rejects short name", func(t *testing.T) {
		svc := NewService(&mockStore{})
		err := svc.EditPlaylist("1", "ab", "", []string{"id1"})
		if err == nil {
			t.Error("expected error for short name, got nil")
		}
	})

	t.Run("rejects long description", func(t *testing.T) {
		svc := NewService(&mockStore{})
		err := svc.EditPlaylist("1", "OK", string(make([]byte, maxDescrLen+1)), []string{"id1"})
		if err == nil {
			t.Error("expected error for long description, got nil")
		}
	})

	t.Run("propagates store error", func(t *testing.T) {
		mock := &mockStore{
			editPlaylistFn: func(id, name, description string, trackIDs []string) error {
				return errors.New("update failed")
			},
		}
		svc := NewService(mock)
		err := svc.EditPlaylist("1", "New Name", "", []string{"id1"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_DeletePlaylist(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		svc := NewService(&mockStore{})
		err := svc.DeletePlaylist("1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockStore{
			deletePlaylistFn: func(id string) error {
				return errors.New("delete failed")
			},
		}
		svc := NewService(mock)
		err := svc.DeletePlaylist("1")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
