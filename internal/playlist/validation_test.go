package playlist

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	t.Run("rejects name too short", func(t *testing.T) {
		err := validateName("ab")
		if err == nil {
			t.Error("expected error for short name, got nil")
		}
		if !strings.Contains(err.Error(), "at least") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		err := validateName("")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("accepts valid name at minimum length", func(t *testing.T) {
		err := validateName("abc")
		if err != nil {
			t.Errorf("expected nil for valid name, got: %v", err)
		}
	})

	t.Run("rejects name too long", func(t *testing.T) {
		longName := strings.Repeat("a", maxNameLen+1)
		err := validateName(longName)
		if err == nil {
			t.Error("expected error for long name, got nil")
		}
		if !strings.Contains(err.Error(), "at most") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("accepts name at maximum length", func(t *testing.T) {
		name := strings.Repeat("a", maxNameLen)
		err := validateName(name)
		if err != nil {
			t.Errorf("expected nil for max-length name, got: %v", err)
		}
	})
}

func TestValidateDescr(t *testing.T) {
	t.Run("accepts empty description", func(t *testing.T) {
		err := validateDescr("")
		if err != nil {
			t.Errorf("expected nil for empty description, got: %v", err)
		}
	})

	t.Run("accepts short description", func(t *testing.T) {
		err := validateDescr("A nice playlist")
		if err != nil {
			t.Errorf("expected nil, got: %v", err)
		}
	})

	t.Run("rejects description too long", func(t *testing.T) {
		longDescr := strings.Repeat("x", maxDescrLen+1)
		err := validateDescr(longDescr)
		if err == nil {
			t.Error("expected error for long description, got nil")
		}
	})

	t.Run("accepts description at maximum length", func(t *testing.T) {
		descr := strings.Repeat("x", maxDescrLen)
		err := validateDescr(descr)
		if err != nil {
			t.Errorf("expected nil for max-length description, got: %v", err)
		}
	})
}

func TestValidateTracks(t *testing.T) {
	t.Run("accepts valid track IDs", func(t *testing.T) {
		err := validateTracks([]string{"id1", "id2", "id3"})
		if err != nil {
			t.Errorf("expected nil, got: %v", err)
		}
	})

	t.Run("rejects empty track ID", func(t *testing.T) {
		err := validateTracks([]string{"id1", "", "id3"})
		if err == nil {
			t.Error("expected error for empty track ID, got nil")
		}
	})

	t.Run("rejects duplicate track IDs", func(t *testing.T) {
		err := validateTracks([]string{"id1", "id1"})
		if err == nil {
			t.Error("expected error for duplicate IDs, got nil")
		}
		if !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("rejects too many tracks", func(t *testing.T) {
		ids := make([]string, maxTracks+1)
		for i := range ids {
			ids[i] = "id"
		}
		err := validateTracks(ids)
		if err == nil {
			t.Error("expected error for too many tracks, got nil")
		}
		if !strings.Contains(err.Error(), "cannot have more than") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("accepts empty track list", func(t *testing.T) {
		err := validateTracks([]string{})
		if err != nil {
			t.Errorf("expected nil for empty track list, got: %v", err)
		}
	})

	t.Run("accepts track list at maximum", func(t *testing.T) {
		ids := make([]string, maxTracks)
		for i := range ids {
			ids[i] = fmt.Sprintf("unique-id-%d", i)
		}
		err := validateTracks(ids)
		if err != nil {
			t.Errorf("expected nil for max tracks, got: %v", err)
		}
	})
}