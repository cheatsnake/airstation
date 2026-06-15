package track

import (
	"math"
	"testing"

	"github.com/cheatsnake/airstation/internal/pkg/hls"
)

func TestRoundDuration(t *testing.T) {
	segDur := float64(hls.DefaultMaxSegmentDuration)

	t.Run("exact multiple stays the same", func(t *testing.T) {
		got := roundDuration(30.0, segDur)
		want := 30.0
		if got != want {
			t.Errorf("roundDuration(30, %v) = %v, want %v", segDur, got, want)
		}
	})

	t.Run("small remainder gets floored", func(t *testing.T) {
		got := roundDuration(31.0, segDur)
		want := 30.0
		if got != want {
			t.Errorf("roundDuration(31, %v) = %v, want %v", segDur, got, want)
		}
	})

	t.Run("remainder less than 1.2 seconds is cropped", func(t *testing.T) {
		got := roundDuration(31.1, segDur)
		want := 30.0
		if got != want {
			t.Errorf("roundDuration(31.1, %v) = %v, want %v", segDur, got, want)
		}
	})

	t.Run("larger remainder floors to whole segment", func(t *testing.T) {
		got := roundDuration(33.0, segDur)
		want := math.Floor(33.0)
		if got != want {
			t.Errorf("roundDuration(33, %v) = %v, want %v", segDur, got, want)
		}
	})

	t.Run("zero duration returns zero", func(t *testing.T) {
		got := roundDuration(0, segDur)
		want := 0.0
		if got != want {
			t.Errorf("roundDuration(0, %v) = %v, want %v", segDur, got, want)
		}
	})

	t.Run("single segment duration", func(t *testing.T) {
		got := roundDuration(5.0, segDur)
		want := 5.0
		if got != want {
			t.Errorf("roundDuration(5, %v) = %v, want %v", segDur, got, want)
		}
	})
}

func TestDefineTrackName(t *testing.T) {
	t.Run("uses metaName when non-empty", func(t *testing.T) {
		got := defineTrackName("file.mp3", "Cool Song")
		want := "Cool Song"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("strips extensions from fileName when metaName is empty", func(t *testing.T) {
		cases := []struct {
			input string
			want  string
		}{
			{"my_song.mp3", "my song"},
			{"my_song.aac", "my song"},
			{"my_song.wav", "my song"},
			{"my_song.flac", "my song"},
		}
		for _, c := range cases {
			got := defineTrackName(c.input, "")
			if got != c.want {
				t.Errorf("defineTrackName(%q, %q) = %q, want %q", c.input, "", got, c.want)
			}
		}
	})

	t.Run("replaces underscores with spaces", func(t *testing.T) {
		got := defineTrackName("artist_song_title.mp3", "")
		want := "artist song title"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("empty metaName and no extension", func(t *testing.T) {
		got := defineTrackName("some_file", "")
		want := "some file"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})
}

func TestReplaceExtension(t *testing.T) {
	t.Run("replaces extension", func(t *testing.T) {
		got := replaceExtension("/path/to/track.mp3", "m4a")
		want := "/path/to/track.m4a"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("adds dot if missing", func(t *testing.T) {
		got := replaceExtension("/path/to/track.mp3", ".m4a")
		want := "/path/to/track.m4a"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("handles empty extension", func(t *testing.T) {
		got := replaceExtension("/path/to/track.mp3", "")
		want := "/path/to/track"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("handles path with no extension", func(t *testing.T) {
		got := replaceExtension("/path/to/track", "m4a")
		want := "/path/to/track.m4a"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})
}