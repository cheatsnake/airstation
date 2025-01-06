package hls

import "testing"

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
