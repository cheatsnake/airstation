package hls

import "fmt"

type Playlist struct {
	SegmentDuration       int // Duration of each segment in the playlist in seconds
	CurrentSegmentPaths   []string
	AvailableSegmentPaths []string
}

func New(duration int, available []string) *Playlist {
	current := make([]string, 0, amountCurrentSegments)
	copy(current, available[0:amountCurrentSegments])

	return &Playlist{
		SegmentDuration:       duration,
		CurrentSegmentPaths:   current,
		AvailableSegmentPaths: available[amountCurrentSegments:],
	}
}

func (hp *Playlist) Generate() string {
	playlist := newHeader(hp.SegmentDuration)

	for _, path := range hp.CurrentSegmentPaths {
		playlist = playlist + newSegment(hp.SegmentDuration, path)
	}

	return playlist
}

func (hp *Playlist) Update(segmentPaths []string) string {
	return ""
}

func newHeader(duration int) string {
	return fmt.Sprintf(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:%d
#EXT-X-MEDIA-SEQUENCE:0`, duration)
}

func newSegment(duration int, path string) string {
	return fmt.Sprintf(`#EXTINF:%d,
%s`, duration, path)
}
