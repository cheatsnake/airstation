package hls

import (
	"testing"
)

func TestNewSegment(t *testing.T) {
	cases := []struct {
		name     string
		duration float64
		path     string
		isFirst  bool
		expected *Segment
	}{
		{
			name:     "Basic segment",
			duration: 5.5,
			path:     "segment0.ts",
			isFirst:  false,
			expected: &Segment{
				Duration: 5.5,
				Path:     "segment0.ts",
				IsFirst:  false,
			},
		},
		{
			name:     "First segment",
			duration: 8.333,
			path:     "segment1.ts",
			isFirst:  true,
			expected: &Segment{
				Duration: 8.333,
				Path:     "segment1.ts",
				IsFirst:  true,
			},
		},
		{
			name:     "Zero duration",
			duration: 0,
			path:     "segment2.ts",
			isFirst:  false,
			expected: &Segment{
				Duration: 0,
				Path:     "segment2.ts",
				IsFirst:  false,
			},
		},
		{
			name:     "Empty path",
			duration: 10.0,
			path:     "",
			isFirst:  false,
			expected: &Segment{
				Duration: 10.0,
				Path:     "",
				IsFirst:  false,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := NewSegment(c.duration, c.path, c.isFirst)

			// Validate the fields of the returned Segment
			if result.Duration != c.expected.Duration {
				t.Errorf("Duration = %f; want %f", result.Duration, c.expected.Duration)
			}
			if result.Path != c.expected.Path {
				t.Errorf("Path = %q; want %q", result.Path, c.expected.Path)
			}
			if result.IsFirst != c.expected.IsFirst {
				t.Errorf("IsFirst = %v; want %v", result.IsFirst, c.expected.IsFirst)
			}
		})
	}
}

func TestGenerateSegments(t *testing.T) {
	cases := []struct {
		name             string
		trackDuration    float64
		segmentDuration  int
		trackID          string
		outDir           string
		expectedSegments []*Segment
	}{
		{
			name:            "Basic case",
			trackDuration:   10.0,
			segmentDuration: 3,
			trackID:         "track1",
			outDir:          "/out",
			expectedSegments: []*Segment{
				{Duration: 3.0, Path: "/out/track10.ts", IsFirst: true},
				{Duration: 3.0, Path: "/out/track11.ts", IsFirst: false},
				{Duration: 3.0, Path: "/out/track12.ts", IsFirst: false},
				{Duration: 1.0, Path: "/out/track13.ts", IsFirst: false},
			},
		},
		{
			name:             "Zero track duration",
			trackDuration:    0,
			segmentDuration:  3,
			trackID:          "track2",
			outDir:           "/out",
			expectedSegments: nil,
		},
		{
			name:             "Zero segment duration",
			trackDuration:    10.0,
			segmentDuration:  0,
			trackID:          "track3",
			outDir:           "/out",
			expectedSegments: nil,
		},
		{
			name:            "Exact division of track duration",
			trackDuration:   9.0,
			segmentDuration: 3,
			trackID:         "track4",
			outDir:          "/out",
			expectedSegments: []*Segment{
				{Duration: 3.0, Path: "/out/track40.ts", IsFirst: true},
				{Duration: 3.0, Path: "/out/track41.ts", IsFirst: false},
				{Duration: 3.0, Path: "/out/track42.ts", IsFirst: false},
			},
		},
		{
			name:            "Large track duration",
			trackDuration:   25.0,
			segmentDuration: 10,
			trackID:         "track5",
			outDir:          "/out",
			expectedSegments: []*Segment{
				{Duration: 10.0, Path: "/out/track50.ts", IsFirst: true},
				{Duration: 10.0, Path: "/out/track51.ts", IsFirst: false},
				{Duration: 5.0, Path: "/out/track52.ts", IsFirst: false},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := GenerateSegments(c.trackDuration, c.segmentDuration, c.trackID, c.outDir)

			if len(result) != len(c.expectedSegments) {
				t.Fatalf("Expected %d segments, got %d", len(c.expectedSegments), len(result))
			}

			for i, seg := range result {
				if seg.Duration != c.expectedSegments[i].Duration {
					t.Errorf("Segment %d: expected duration %f, got %f", i, c.expectedSegments[i].Duration, seg.Duration)
				}
				if seg.Path != c.expectedSegments[i].Path {
					t.Errorf("Segment %d: expected path %q, got %q", i, c.expectedSegments[i].Path, seg.Path)
				}
				if seg.IsFirst != c.expectedSegments[i].IsFirst {
					t.Errorf("Segment %d: expected IsFirst %v, got %v", i, c.expectedSegments[i].IsFirst, seg.IsFirst)
				}
			}
		})
	}
}
