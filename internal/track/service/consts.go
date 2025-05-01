package trackservice

import "github.com/cheatsnake/airstation/internal/hls"

const (
	minAllowedTrackDuration = hls.DefaultMaxSegmentDuration * hls.DefaultLiveSegmentsAmount
	maxAllowedTrackDuration = 6000 // 100 min (just an adequate barrier)
)
