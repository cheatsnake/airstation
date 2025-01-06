package hls

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestNewPlaylist(t *testing.T) {
	current := []Segment{{Duration: 10.5, Path: "segment1.ts"}}
	next := []Segment{{Duration: 9.0, Path: "segment2.ts"}}
	maxDuration := 10
	liveAmount := 2
	playlist := NewPlaylist(current, next, maxDuration, liveAmount)

	if playlist.maxSegmentDuration != maxDuration {
		t.Errorf("Expected maxSegmentDuration to be %d, got %d", maxDuration, playlist.maxSegmentDuration)
	}

	if playlist.liveSegmentsAmount != liveAmount {
		t.Errorf("Expected liveSegmentsAmount to be %d, got %d", liveAmount, playlist.liveSegmentsAmount)
	}

	if len(playlist.currentTrackSegments) != len(current) {
		t.Errorf("Expected %d segment in currentTrackSegments, got %d", len(current), len(playlist.currentTrackSegments))
	}

	if len(playlist.nextTrackSegments) != len(next) {
		t.Errorf("Expected %d segment in nextTrackSegments, got %d", len(next), len(playlist.nextTrackSegments))
	}
}

func TestGenerate(t *testing.T) {
	cases := []struct {
		name          string
		current       []Segment
		next          []Segment
		elapsedTime   int
		liveAmount    int
		expectedPaths []string
		unexpected    []string
	}{
		{
			name: "full from current track",
			current: []Segment{
				{Duration: 10.0, Path: "segment1.ts"},
				{Duration: 10.0, Path: "segment2.ts"},
				{Duration: 10.0, Path: "segment3.ts"},
			},
			next: []Segment{
				{Duration: 10.0, Path: "segment4.ts"},
			},
			elapsedTime:   0,
			liveAmount:    3,
			expectedPaths: []string{"segment1.ts", "segment2.ts", "segment3.ts"},
			unexpected:    []string{"segment4.ts"},
		},
		{
			name: "partial from current and next track",
			current: []Segment{
				{Duration: 10.0, Path: "segment1.ts"},
				{Duration: 10.0, Path: "segment2.ts"},
			},
			next: []Segment{
				{Duration: 10.0, Path: "segment3.ts"},
				{Duration: 10.0, Path: "segment4.ts"},
			},
			elapsedTime:   10,
			liveAmount:    3,
			expectedPaths: []string{"segment2.ts", "segment3.ts", "segment4.ts"},
			unexpected:    []string{"segment1.ts"},
		},
		{
			name: "full from next track",
			current: []Segment{
				{Duration: 10.0, Path: "segment1.ts"},
			},
			next: []Segment{
				{Duration: 10.0, Path: "segment2.ts"},
				{Duration: 10.0, Path: "segment3.ts"},
				{Duration: 10.0, Path: "segment4.ts"},
			},
			elapsedTime:   20,
			liveAmount:    3,
			expectedPaths: []string{"segment2.ts", "segment3.ts", "segment4.ts"},
			unexpected:    []string{"segment1.ts"},
		},
		{
			name: "not enough segments",
			current: []Segment{
				{Duration: 10.0, Path: "segment1.ts"},
			},
			next: []Segment{
				{Duration: 10.0, Path: "segment2.ts"},
			},
			elapsedTime:   0,
			liveAmount:    5,
			expectedPaths: []string{"segment1.ts", "segment2.ts"},
			unexpected:    []string{"segment3.ts"},
		},
		{
			name:          "empty tracks",
			current:       []Segment{},
			next:          []Segment{},
			elapsedTime:   0,
			liveAmount:    3,
			expectedPaths: []string{},
			unexpected:    []string{"segment1.ts"},
		},
		{
			name: "start index beyond current track",
			current: []Segment{
				{Duration: 10.0, Path: "segment1.ts"},
			},
			next: []Segment{
				{Duration: 10.0, Path: "segment2.ts"},
				{Duration: 10.0, Path: "segment3.ts"}},
			elapsedTime:   20,
			liveAmount:    3,
			expectedPaths: []string{"segment2.ts", "segment3.ts"},
			unexpected:    []string{"segment1.ts"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			playlist := NewPlaylist(c.current, c.next, 10, c.liveAmount)
			result := playlist.Generate(c.elapsedTime)

			for _, path := range c.expectedPaths {
				if !strings.Contains(result, path) {
					t.Errorf("Expected segment %s not found in playlist: %s", path, result)
				}
			}

			for _, path := range c.unexpected {
				if strings.Contains(result, path) {
					t.Errorf("Unexpected segment %s found in playlist: %s", path, result)
				}
			}

			if !strings.HasPrefix(result, "#EXTM3U") {
				t.Errorf("Playlist header is missing: %s", result)
			}
		})
	}
}

func TestNext(t *testing.T) {
	current := []Segment{
		{Duration: 10.0, Path: "segment1.ts"},
		{Duration: 10.0, Path: "segment2.ts"},
	}
	next := []Segment{
		{Duration: 10.0, Path: "segment3.ts"},
	}
	maxDuration := 10
	liveAmount := 2

	playlist := NewPlaylist(current, next, maxDuration, liveAmount)

	playlist.Next([]Segment{{Duration: 8.0, Path: "segment4.ts"}})

	if len(playlist.currentTrackSegments) != 1 || playlist.currentTrackSegments[0].Path != "segment3.ts" {
		t.Errorf("Expected currentTrackSegments to contain nextTrackSegments, got: %v", playlist.currentTrackSegments)
	}

	if len(playlist.nextTrackSegments) != 1 || playlist.nextTrackSegments[0].Path != "segment4.ts" {
		t.Errorf("Expected nextTrackSegments to be updated, got: %v", playlist.nextTrackSegments)
	}
}

func TestAddSegments(t *testing.T) {
	current := []Segment{{Duration: 10.0, Path: "segment1.ts"}}
	next := []Segment{{Duration: 10.0, Path: "segment2.ts"}}
	maxDuration := 10
	liveAmount := 2
	playlist := NewPlaylist(current, next, maxDuration, liveAmount)

	newSegments := []Segment{
		{Duration: 8.0, Path: "segment3.ts"},
		{Duration: 7.5, Path: "segment4.ts"},
	}
	playlist.AddSegments(newSegments)

	if len(playlist.nextTrackSegments) != 3 {
		t.Errorf("Expected nextTrackSegments to contain 3 segments, got: %d", len(playlist.nextTrackSegments))
	}

	if playlist.nextTrackSegments[2].Path != "segment4.ts" {
		t.Errorf("Expected last segment to be segment4.ts, got: %s", playlist.nextTrackSegments[2].Path)
	}
}

func TestCollectLiveSegments(t *testing.T) {
	t.Run("full from current track", func(t *testing.T) {
		current := []Segment{
			{Duration: 10.0, Path: "segment1.ts"},
			{Duration: 10.0, Path: "segment2.ts"},
			{Duration: 10.0, Path: "segment3.ts"},
		}
		next := []Segment{
			{Duration: 10.0, Path: "segment4.ts"},
		}

		playlist := NewPlaylist(current, next, 10, 3)
		liveSegments := playlist.collectLiveSegments(0)
		expected := current

		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("partial from current and next track", func(t *testing.T) {
		current := []Segment{
			{Duration: 10.0, Path: "segment1.ts"},
			{Duration: 10.0, Path: "segment2.ts"},
		}
		next := []Segment{
			{Duration: 10.0, Path: "segment3.ts"},
			{Duration: 10.0, Path: "segment4.ts"},
		}
		playlist := NewPlaylist(current, next, 10, 3)

		liveSegments := playlist.collectLiveSegments(1)

		expected := []Segment{
			{Duration: 10.0, Path: "segment2.ts"},
			{Duration: 10.0, Path: "segment3.ts"},
			{Duration: 10.0, Path: "segment4.ts"},
		}
		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("full from next track", func(t *testing.T) {
		current := []Segment{
			{Duration: 10.0, Path: "segment1.ts"},
		}
		next := []Segment{
			{Duration: 10.0, Path: "segment2.ts"},
			{Duration: 10.0, Path: "segment3.ts"},
			{Duration: 10.0, Path: "segment4.ts"},
		}
		playlist := NewPlaylist(current, next, 10, 3)

		liveSegments := playlist.collectLiveSegments(2)
		expected := next

		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("not enough segments", func(t *testing.T) {
		current := []Segment{
			{Duration: 10.0, Path: "segment1.ts"},
		}
		next := []Segment{
			{Duration: 10.0, Path: "segment2.ts"},
		}
		playlist := NewPlaylist(current, next, 10, 5)

		liveSegments := playlist.collectLiveSegments(0)

		expected := []Segment{
			{Duration: 10.0, Path: "segment1.ts"},
			{Duration: 10.0, Path: "segment2.ts"},
		}
		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("start index beyond current track", func(t *testing.T) {
		current := []Segment{
			{Duration: 10.0, Path: "segment1.ts"},
		}
		next := []Segment{
			{Duration: 10.0, Path: "segment2.ts"},
			{Duration: 10.0, Path: "segment3.ts"},
		}

		playlist := NewPlaylist(current, next, 10, 3)
		liveSegments := playlist.collectLiveSegments(2)
		expected := next

		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("empty tracks", func(t *testing.T) {
		current := []Segment{}
		next := []Segment{}
		playlist := NewPlaylist(current, next, 10, 3)

		liveSegments := playlist.collectLiveSegments(0)

		expected := []Segment{}
		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})
}

func TestHLSHeader(t *testing.T) {
	cases := []int{1, 5, 10, 20}

	for _, c := range cases {
		header := hlsHeader(c)
		if !strings.Contains(header, "#EXT-X-TARGETDURATION:"+strconv.Itoa(c)) {
			t.Errorf("HLS header missing expected target duration: %s", header)
		}
	}
}

func TestHLSSegment(t *testing.T) {
	cases := []Segment{
		{Duration: 1, Path: "seg1.ts"},
		{Duration: 5, Path: "path/to/seg2.ts"},
		{Duration: 7.5, Path: "seg3.ts"},
		{Duration: 12.25, Path: "seg4.ts"},
		{Duration: 15, Path: "seg5.ts"},
	}

	for _, c := range cases {
		segment := hlsSegment(c.Duration, c.Path)
		if !strings.Contains(segment, c.Path) ||
			!strings.Contains(segment, "#EXTINF:"+strconv.FormatFloat(c.Duration, 'f', -1, 64)) {
			t.Errorf("HLS segment does not match expected format: %s", segment)
		}
	}
}
