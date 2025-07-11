package track

import "github.com/cheatsnake/airstation/internal/hls"

const (
	minAllowedTrackDuration = hls.DefaultMaxSegmentDuration * hls.DefaultLiveSegmentsAmount
	maxAllowedTrackDuration = 6000 // 100 min (just an adequate barrier)
	defaultAudioBitRate     = 192  // best balance between quallity and size
)

const (
	m4aExtension = "m4a"
	mp3Extension = "mp3"
	aacExtension = "aac"
	wavExtension = "wav"
)
