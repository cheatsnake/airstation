package hls

import (
	"math"
	"path/filepath"
	"strconv"
)

type Segment struct {
	Duration float64
	Path     string
}

func NewSegment(duration float64, path string) *Segment {
	return &Segment{
		Duration: duration,
		Path:     path,
	}
}

func GenerateSegments(trackDuration, segmentDuration float64, trackID, outDir string) []*Segment {
	if trackDuration <= 0 || segmentDuration <= 0 {
		return nil
	}

	// Calculate total possible number of segments (rounded up)
	totalSegments := int(math.Round(trackDuration / segmentDuration))
	segments := make([]*Segment, 0, totalSegments)

	remaining := trackDuration
	index := 0

	// Generate segments until the entire track is covered
	for remaining > 0 {
		segName := trackID + strconv.Itoa(index) + segmentExtension
		segPath := filepath.Join(outDir, segName)

		// Use the smaller of the remaining or full segment duration
		duration := math.Min(remaining, segmentDuration)
		segments = append(segments, NewSegment(duration, segPath))

		remaining -= duration
		index++
	}

	return segments
}
