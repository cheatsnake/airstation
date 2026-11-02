package track

import "github.com/cheatsnake/airstation/internal/pkg/hls"

const (
	minAllowedTrackDuration = hls.DefaultMaxSegmentDuration * hls.DefaultLiveSegmentsAmount
	// Jingles, ads, station IDs and voice links are usually much shorter than
	// a song, so they only need to be long enough to form one HLS segment.
	minAllowedAssetDuration = 1.0
	maxAllowedTrackDuration = 36000 // 10 hours (just an adequate barrier)
	defaultAudioBitRate     = 48    // stereo default; operator can raise it in Studio
)

const (
	m4aExtension  = "m4a"
	mp3Extension  = "mp3"
	aacExtension  = "aac"
	wavExtension  = "wav"
	flacExtension = "flac"
)

const (
	KindMusic     = "music"
	KindJingle    = "jingle"
	KindAd        = "ad"
	KindStationID = "station_id"
	KindVoice     = "voice"
)

var kinds = map[string]struct{}{
	KindMusic:     {},
	KindJingle:    {},
	KindAd:        {},
	KindStationID: {},
	KindVoice:     {},
}

// IsValidKind reports whether kind is a supported asset type.
func IsValidKind(kind string) bool {
	_, ok := kinds[kind]
	return ok
}

// IsInterlude reports whether the kind is a non-music asset that can be
// inserted into the queue by the scheduler.
func IsInterlude(kind string) bool {
	switch kind {
	case KindJingle, KindAd, KindStationID:
		return true
	default:
		return false
	}
}
