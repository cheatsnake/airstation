package sse

import "testing"

func TestNewEvent(t *testing.T) {
	t.Run("sets name and data", func(t *testing.T) {
		e := NewEvent("play", "track1")
		if e.Name != "play" {
			t.Errorf("expected Name %q, got %q", "play", e.Name)
		}
		if e.Data != "track1" {
			t.Errorf("expected Data %q, got %q", "track1", e.Data)
		}
	})
}

func TestEvent_Stringify(t *testing.T) {
	t.Run("formats SSE with both fields", func(t *testing.T) {
		e := NewEvent("play", "track1")
		got := e.Stringify()
		if want := "event: play\n"; got[:len(want)] != want {
			t.Errorf("missing event line, got: %q", got)
		}
		if want := "data: track1\n"; !contains(got, want) {
			t.Errorf("missing data line, got: %q", got)
		}
		if !endsWith(got, "\n\n") {
			t.Errorf("expected double newline at end, got: %q", got)
		}
	})

	t.Run("omits event line when name is empty", func(t *testing.T) {
		e := NewEvent("", "payload")
		got := e.Stringify()
		if contains(got, "event:") {
			t.Errorf("unexpected event line in: %q", got)
		}
		if !contains(got, "data: payload\n") {
			t.Errorf("missing data line in: %q", got)
		}
	})

	t.Run("omits data line when data is empty", func(t *testing.T) {
		e := NewEvent("ping", "")
		got := e.Stringify()
		if !contains(got, "event: ping\n") {
			t.Errorf("missing event line in: %q", got)
		}
		if contains(got, "data:") {
			t.Errorf("unexpected data line in: %q", got)
		}
	})

	t.Run("both empty still has double newline", func(t *testing.T) {
		e := NewEvent("", "")
		got := e.Stringify()
		if got != "\n" {
			t.Errorf("expected just newline for empty event, got: %q", got)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}