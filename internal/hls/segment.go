package hls

import (
	"math"
	"path/filepath"
	"strconv"
)

// Segment represents a single segment in an HLS playlist.
type Segment struct {
	Duration float64 // The length of the segment in seconds.
	Path     string  // The file path or URL of the segment.
	IsFirst  bool    // A flag indicating whether this segment is the first segment in the track.
}

// NewSegment creates and returns a new Segment instance with the provided duration, path, and first segment flag.
//
// Parameters:
//   - duration: The duration of the segment in seconds.
//   - path: The file path or URL of the segment.
//   - isFirst: A boolean flag indicating whether this segment is the first segment in the track.
//
// Returns:
//   - A pointer to the newly created Segment instance.
func NewSegment(duration float64, path string, isFirst bool) *Segment {
	return &Segment{
		Duration: duration,
		Path:     path,
		IsFirst:  isFirst,
	}
}

// GenerateSegments creates a list of Segment instances for a given track based on its duration and segment duration.
// It divides the track into segments of the specified duration and generates metadata for each segment.
//
// Parameters:
//   - trackDuration: The total duration of the track in seconds.
//   - segmentDuration: The desired duration of each segment in seconds.
//   - trackID: The unique identifier for the track, used to generate segment names.
//   - outDir: The output directory where the segments will be stored.
//
// Returns:
//   - A slice of pointers to Segment instances representing the generated segments.
func GenerateSegments(trackDuration float64, segmentDuration int, trackID, outDir string) []*Segment {
	if trackDuration <= 0 || segmentDuration <= 0 {
		return []*Segment{}
	}

	// Calculate total possible number of segments (rounded up)
	totalSegments := int(math.Round(trackDuration / float64(segmentDuration)))
	segments := make([]*Segment, 0, totalSegments)

	remaining := trackDuration
	index := 0

	// Generate segments until the entire track is covered
	for remaining > 0 {
		segName := trackID + strconv.Itoa(index) + SegmentExtension
		segPath := filepath.Join(outDir, segName)

		// Use the smaller of the remaining or full segment duration
		duration := math.Min(remaining, float64(segmentDuration))
		isFirst := index == 0
		segments = append(segments, NewSegment(duration, segPath, isFirst))

		remaining -= duration
		index++
	}

	return segments
}
