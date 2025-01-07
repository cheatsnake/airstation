package hls

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewSegment(t *testing.T) {
	cases := []Segment{
		{Duration: 1, Path: "path/to/segment.ts"},
		{Duration: 5, Path: "segment.ts"},
		{Duration: 10, Path: "path/to/segment.ts"},
		{Duration: 10.5, Path: "path/to/segment.ts"},
		{Duration: 0.5, Path: "path/to/segment.ts"},
	}

	for _, c := range cases {
		segment := NewSegment(c.Duration, c.Path)
		if segment.Duration != c.Duration {
			t.Errorf("Expected duration to be %f, got %f", c.Duration, segment.Duration)
		}
		if segment.Path != c.Path {
			t.Errorf("Expected path to be '%s', got %s", c.Path, segment.Path)
		}
	}
}

func TestGenerateSegments(t *testing.T) {
	cases := []struct {
		name            string
		trackDuration   float64
		segmentDuration float64
		trackID         string
		outDir          string
		expected        []*Segment
	}{
		{
			name:            "Exact division",
			trackDuration:   10.0,
			segmentDuration: 2.0,
			trackID:         "track1",
			outDir:          "/tmp",
			expected: []*Segment{
				{Duration: 2.0, Path: filepath.Join("/tmp", "track10.ts")},
				{Duration: 2.0, Path: filepath.Join("/tmp", "track11.ts")},
				{Duration: 2.0, Path: filepath.Join("/tmp", "track12.ts")},
				{Duration: 2.0, Path: filepath.Join("/tmp", "track13.ts")},
				{Duration: 2.0, Path: filepath.Join("/tmp", "track14.ts")},
			},
		},
		{
			name:            "Remainder segment",
			trackDuration:   9.5,
			segmentDuration: 2.0,
			trackID:         "track2",
			outDir:          "/tmp",
			expected: []*Segment{
				{Duration: 2.0, Path: filepath.Join("/tmp", "track20.ts")},
				{Duration: 2.0, Path: filepath.Join("/tmp", "track21.ts")},
				{Duration: 2.0, Path: filepath.Join("/tmp", "track22.ts")},
				{Duration: 2.0, Path: filepath.Join("/tmp", "track23.ts")},
				{Duration: 1.5, Path: filepath.Join("/tmp", "track24.ts")},
			},
		},
		{
			name:            "Track duration less than segment duration",
			trackDuration:   1.5,
			segmentDuration: 2.0,
			trackID:         "track3",
			outDir:          "/tmp",
			expected: []*Segment{
				{Duration: 1.5, Path: filepath.Join("/tmp", "track30.ts")},
			},
		},
		{
			name:            "Track duration is zero",
			trackDuration:   0.0,
			segmentDuration: 2.0,
			trackID:         "track4",
			outDir:          "/tmp",
			expected:        nil,
		},
		{
			name:            "Segment duration is zero",
			trackDuration:   10.0,
			segmentDuration: 0.0,
			trackID:         "track5",
			outDir:          "/tmp",
			expected:        nil,
		},
		{
			name:            "Negative track duration",
			trackDuration:   -5.0,
			segmentDuration: 2.0,
			trackID:         "track6",
			outDir:          "/tmp",
			expected:        nil,
		},
		{
			name:            "Negative segment duration",
			trackDuration:   10.0,
			segmentDuration: -2.0,
			trackID:         "track7",
			outDir:          "/tmp",
			expected:        nil,
		},
		{
			name:            "Negative track and segment duration",
			trackDuration:   -3.0,
			segmentDuration: -2.0,
			trackID:         "track7",
			outDir:          "/tmp",
			expected:        nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := GenerateSegments(c.trackDuration, c.segmentDuration, c.trackID, c.outDir)

			// Compare the results
			if !reflect.DeepEqual(got, c.expected) {
				t.Errorf("GenerateSegments() = %v, expected %v", got, c.expected)
			}
		})
	}
}
