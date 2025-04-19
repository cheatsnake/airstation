package hls

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewPlaylist(t *testing.T) {
	current := []*Segment{{Duration: 10.5, Path: "segment1.ts"}}
	next := []*Segment{{Duration: 9.0, Path: "segment2.ts"}}
	playlist := NewPlaylist(current, next)

	if playlist.MaxSegmentDuration != DefaultMaxSegmentDuration {
		t.Errorf("Expected maxSegmentDuration to be %d, got %d", DefaultMaxSegmentDuration, playlist.MaxSegmentDuration)
	}

	if playlist.LiveSegmentsAmount != DefaultLiveSegmentsAmount {
		t.Errorf("Expected liveSegmentsAmount to be %d, got %d", DefaultLiveSegmentsAmount, playlist.LiveSegmentsAmount)
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
		current       []*Segment
		next          []*Segment
		elapsedTime   float64
		expectedPaths []string
		unexpected    []string
	}{
		{
			name: "full from current track",
			current: []*Segment{
				{Duration: 5.0, Path: "segment1.ts"},
				{Duration: 5.0, Path: "segment2.ts"},
				{Duration: 5.0, Path: "segment3.ts"},
			},
			next: []*Segment{
				{Duration: 5.0, Path: "segment4.ts"},
			},
			elapsedTime:   0.0,
			expectedPaths: []string{"segment1.ts", "segment2.ts", "segment3.ts"},
			unexpected:    []string{"segment4.ts"},
		},
		{
			name: "partial from current and next track",
			current: []*Segment{
				{Duration: 5.0, Path: "segment1.ts"},
				{Duration: 5.0, Path: "segment2.ts"},
			},
			next: []*Segment{
				{Duration: 5.0, Path: "segment3.ts"},
				{Duration: 5.0, Path: "segment4.ts"},
			},
			elapsedTime:   5.0,
			expectedPaths: []string{"segment2.ts", "segment3.ts", "segment4.ts"},
			unexpected:    []string{"segment1.ts"},
		},
		{
			name: "full from next track",
			current: []*Segment{
				{Duration: 5.0, Path: "segment1.ts"},
			},
			next: []*Segment{
				{Duration: 5.0, Path: "segment2.ts"},
				{Duration: 5.0, Path: "segment3.ts"},
				{Duration: 5.0, Path: "segment4.ts"},
			},
			elapsedTime:   20.0,
			expectedPaths: []string{"segment2.ts", "segment3.ts", "segment4.ts"},
			unexpected:    []string{"segment1.ts"},
		},
		{
			name: "not enough segments",
			current: []*Segment{
				{Duration: 5.0, Path: "segment1.ts"},
			},
			next: []*Segment{
				{Duration: 5.0, Path: "segment2.ts"},
			},
			elapsedTime:   0.0,
			expectedPaths: []string{"segment1.ts", "segment2.ts"},
			unexpected:    []string{"segment3.ts"},
		},
		{
			name:          "empty tracks",
			current:       []*Segment{},
			next:          []*Segment{},
			elapsedTime:   0.0,
			expectedPaths: []string{},
			unexpected:    []string{"segment1.ts"},
		},
		{
			name: "start index beyond current track",
			current: []*Segment{
				{Duration: 5.0, Path: "segment1.ts"},
			},
			next: []*Segment{
				{Duration: 5.0, Path: "segment2.ts"},
				{Duration: 5.0, Path: "segment3.ts"}},
			elapsedTime:   20.0,
			expectedPaths: []string{"segment2.ts", "segment3.ts"},
			unexpected:    []string{"segment1.ts"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			playlist := NewPlaylist(c.current, c.next)
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
	current := []*Segment{
		{Duration: 5.0, Path: "segment1.ts"},
		{Duration: 5.0, Path: "segment2.ts"},
	}
	next := []*Segment{
		{Duration: 5.0, Path: "segment3.ts"},
	}

	playlist := NewPlaylist(current, next)

	playlist.Next([]*Segment{{Duration: 8.0, Path: "segment4.ts"}})

	if len(playlist.currentTrackSegments) != 1 || playlist.currentTrackSegments[0].Path != "segment3.ts" {
		t.Errorf("Expected currentTrackSegments to contain nextTrackSegments, got: %v", playlist.currentTrackSegments)
	}

	if len(playlist.nextTrackSegments) != 1 || playlist.nextTrackSegments[0].Path != "segment4.ts" {
		t.Errorf("Expected nextTrackSegments to be updated, got: %v", playlist.nextTrackSegments)
	}
}

func TestAddSegments(t *testing.T) {
	current := []*Segment{{Duration: 5.0, Path: "segment1.ts"}}
	next := []*Segment{{Duration: 5.0, Path: "segment2.ts"}}
	playlist := NewPlaylist(current, next)

	newSegments := []*Segment{
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
		current := []*Segment{
			{Duration: 5.0, Path: "segment1.ts"},
			{Duration: 5.0, Path: "segment2.ts"},
			{Duration: 5.0, Path: "segment3.ts"},
		}
		next := []*Segment{
			{Duration: 5.0, Path: "segment4.ts"},
		}

		playlist := NewPlaylist(current, next)
		liveSegments := playlist.currentSegments(0)
		expected := current

		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("partitial from current track", func(t *testing.T) {
		current := []*Segment{
			{Duration: 5.0, Path: "segment1.ts"},
			{Duration: 5.0, Path: "segment2.ts"},
			{Duration: 5.0, Path: "segment3.ts"},
		}
		next := []*Segment{
			{Duration: 5.0, Path: "segment4.ts"},
		}

		playlist := NewPlaylist(current, next)
		liveSegments := playlist.currentSegments(0)
		expected := []*Segment{
			{Duration: 5.0, Path: "segment1.ts"},
			{Duration: 5.0, Path: "segment2.ts"},
			{Duration: 5.0, Path: "segment3.ts"},
		}

		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("partial from current and next track", func(t *testing.T) {
		current := []*Segment{
			{Duration: 5.0, Path: "segment1.ts"},
			{Duration: 5.0, Path: "segment2.ts"},
		}
		next := []*Segment{
			{Duration: 5.0, Path: "segment3.ts"},
			{Duration: 5.0, Path: "segment4.ts"},
		}
		playlist := NewPlaylist(current, next)

		liveSegments := playlist.currentSegments(1)

		expected := []*Segment{
			{Duration: 5.0, Path: "segment2.ts"},
			{Duration: 5.0, Path: "segment3.ts"},
			{Duration: 5.0, Path: "segment4.ts"},
		}
		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("full from next track", func(t *testing.T) {
		current := []*Segment{
			{Duration: 5.0, Path: "segment1.ts"},
		}
		next := []*Segment{
			{Duration: 5.0, Path: "segment2.ts"},
			{Duration: 5.0, Path: "segment3.ts"},
			{Duration: 5.0, Path: "segment4.ts"},
		}
		playlist := NewPlaylist(current, next)

		liveSegments := playlist.currentSegments(2)
		expected := next

		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("not enough segments", func(t *testing.T) {
		current := []*Segment{
			{Duration: 5.0, Path: "segment1.ts"},
		}
		next := []*Segment{
			{Duration: 5.0, Path: "segment2.ts"},
		}
		playlist := NewPlaylist(current, next)

		liveSegments := playlist.currentSegments(0)

		expected := []*Segment{
			{Duration: 5.0, Path: "segment1.ts"},
			{Duration: 5.0, Path: "segment2.ts"},
		}
		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("start index beyond current track", func(t *testing.T) {
		current := []*Segment{
			{Duration: 5.0, Path: "segment1.ts"},
		}
		next := []*Segment{
			{Duration: 5.0, Path: "segment2.ts"},
			{Duration: 5.0, Path: "segment3.ts"},
		}

		playlist := NewPlaylist(current, next)
		liveSegments := playlist.currentSegments(2)
		expected := next

		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})

	t.Run("empty tracks", func(t *testing.T) {
		current := []*Segment{}
		next := []*Segment{}
		playlist := NewPlaylist(current, next)

		liveSegments := playlist.currentSegments(0)

		expected := []*Segment{}
		if !reflect.DeepEqual(liveSegments, expected) {
			t.Errorf("Expected %v, got %v", expected, liveSegments)
		}
	})
}

func TestHlsHeader(t *testing.T) {
	cases := []struct {
		name      string
		dur       int
		mediaSeq  int64
		disconSeq int64
		offset    float64
		expected  string
	}{
		{
			name:      "Basic case",
			dur:       10,
			mediaSeq:  1,
			disconSeq: 0,
			offset:    0.0,
			expected: "#EXTM3U\n" +
				"#EXT-X-VERSION:6\n" +
				"#EXT-X-TARGETDURATION:10\n" +
				"#EXT-X-MEDIA-SEQUENCE:1\n" +
				"#EXT-X-DISCONTINUITY-SEQUENCE:0\n" +
				"#EXT-X-START:TIME-OFFSET=0.00\n",
		},
		{
			name:      "Non-zero discontinuity sequence",
			dur:       15,
			mediaSeq:  5,
			disconSeq: 2,
			offset:    5.5,
			expected: "#EXTM3U\n" +
				"#EXT-X-VERSION:6\n" +
				"#EXT-X-TARGETDURATION:15\n" +
				"#EXT-X-MEDIA-SEQUENCE:5\n" +
				"#EXT-X-DISCONTINUITY-SEQUENCE:2\n" +
				"#EXT-X-START:TIME-OFFSET=5.50\n",
		},
		{
			name:      "Zero values",
			dur:       0,
			mediaSeq:  0,
			disconSeq: 0,
			offset:    0.0,
			expected: "#EXTM3U\n" +
				"#EXT-X-VERSION:6\n" +
				"#EXT-X-TARGETDURATION:0\n" +
				"#EXT-X-MEDIA-SEQUENCE:0\n" +
				"#EXT-X-DISCONTINUITY-SEQUENCE:0\n" +
				"#EXT-X-START:TIME-OFFSET=0.00\n",
		},
		{
			name:      "Large values",
			dur:       999,
			mediaSeq:  123456789,
			disconSeq: 987654321,
			offset:    123.45,
			expected: "#EXTM3U\n" +
				"#EXT-X-VERSION:6\n" +
				"#EXT-X-TARGETDURATION:999\n" +
				"#EXT-X-MEDIA-SEQUENCE:123456789\n" +
				"#EXT-X-DISCONTINUITY-SEQUENCE:987654321\n" +
				"#EXT-X-START:TIME-OFFSET=123.45\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := hlsHeader(c.dur, c.mediaSeq, c.disconSeq, c.offset)
			if result != c.expected {
				t.Errorf("hlsHeader(%d, %d, %d, %.2f) = %q; want %q", c.dur, c.mediaSeq, c.disconSeq, c.offset, result, c.expected)
			}
		})
	}
}

func TestHlsSegment(t *testing.T) {
	cases := []struct {
		name     string
		dur      float64
		path     string
		isDiscon bool
		expected string
	}{
		{
			name:     "Basic segment without discontinuity",
			dur:      5.5,
			path:     "segment0.ts",
			isDiscon: false,
			expected: "#EXTINF:5.50,\nsegment0.ts\n",
		},
		{
			name:     "Segment with discontinuity",
			dur:      8.333,
			path:     "segment1.ts",
			isDiscon: true,
			expected: "#EXT-X-DISCONTINUITY\n#EXTINF:8.33,\nsegment1.ts\n",
		},
		{
			name:     "Zero duration segment without discontinuity",
			dur:      0,
			path:     "segment2.ts",
			isDiscon: false,
			expected: "#EXTINF:0.00,\nsegment2.ts\n",
		},
		{
			name:     "Large duration segment with discontinuity",
			dur:      1234.56789,
			path:     "segment3.ts",
			isDiscon: true,
			expected: "#EXT-X-DISCONTINUITY\n#EXTINF:1234.57,\nsegment3.ts\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := hlsSegment(c.dur, c.path, c.isDiscon)
			if result != c.expected {
				t.Errorf("hlsSegment(%f, %q, %v) = %q; want %q", c.dur, c.path, c.isDiscon, result, c.expected)
			}
		})
	}
}
