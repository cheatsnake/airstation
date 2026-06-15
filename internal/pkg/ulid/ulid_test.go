package ulid

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("returns non-empty string of correct length", func(t *testing.T) {
		got := New()
		if len(got) != Length {
			t.Errorf("expected length %d, got %d", Length, len(got))
		}
	})

	t.Run("returns lowercase", func(t *testing.T) {
		got := New()
		if got != strings.ToLower(got) {
			t.Errorf("expected lowercase, got %q", got)
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		ids := make(map[string]struct{}, 100)
		for i := 0; i < 100; i++ {
			id := New()
			if _, exists := ids[id]; exists {
				t.Errorf("duplicate ID generated: %s", id)
			}
			ids[id] = struct{}{}
		}
	})
}

func TestVerify(t *testing.T) {
	t.Run("valid ULID passes", func(t *testing.T) {
		id := New()
		if err := Verify(id); err != nil {
			t.Errorf("expected nil error for valid ULID, got: %v", err)
		}
	})

	t.Run("invalid string returns error", func(t *testing.T) {
		err := Verify("not-a-valid-id")
		if err == nil {
			t.Error("expected error for invalid ID, got nil")
		}
	})

	t.Run("empty string returns error", func(t *testing.T) {
		err := Verify("")
		if err == nil {
			t.Error("expected error for empty string, got nil")
		}
	})

	t.Run("truncated ULID returns error", func(t *testing.T) {
		id := New()
		err := Verify(id[:10])
		if err == nil {
			t.Error("expected error for truncated ULID, got nil")
		}
	})
}