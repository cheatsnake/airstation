package hls

import (
	"math"
	"strconv"
)

type Playlist struct {
	liveSegmentsAmount   int
	maxSegmentDuration   float64
	currentTrackSegments []*Segment
	nextTrackSegments    []*Segment
}

func NewPlaylist(cur, next []*Segment, maxDuration float64, liveAmount int) *Playlist {
	return &Playlist{
		liveSegmentsAmount:   liveAmount,
		maxSegmentDuration:   maxDuration,
		currentTrackSegments: cur,
		nextTrackSegments:    next,
	}
}

func (p *Playlist) Generate(elapsedTime float64) string {
	playlist := hlsHeader(p.maxSegmentDuration)
	firstSegmentIndex := int(math.Floor(elapsedTime / p.maxSegmentDuration))
	liveSegments := p.collectLiveSegments(firstSegmentIndex)

	for _, seg := range liveSegments {
		playlist += hlsSegment(seg.Duration, seg.Path)
	}

	return playlist
}

func (p *Playlist) Next(next []*Segment) {
	p.currentTrackSegments = p.nextTrackSegments
	p.nextTrackSegments = next
}

func (p *Playlist) AddSegments(segments []*Segment) {
	p.nextTrackSegments = append(p.nextTrackSegments, segments...)
}

// collectLiveSegments gathers enough segments from current and next tracks to meet liveSegmentsAmount
func (p *Playlist) collectLiveSegments(startIndex int) []*Segment {
	liveSegments := make([]*Segment, 0, p.liveSegmentsAmount)

	if startIndex < len(p.currentTrackSegments) {
		endIndex := startIndex + p.liveSegmentsAmount
		if endIndex >= len(p.currentTrackSegments) {
			endIndex = len(p.currentTrackSegments)
		}

		liveSegments = append(liveSegments, p.currentTrackSegments[startIndex:endIndex]...)
	}

	if len(liveSegments) < p.liveSegmentsAmount {
		required := p.liveSegmentsAmount - len(liveSegments)
		liveSegments = append(liveSegments, p.nextTrackSegments[:min(len(p.nextTrackSegments), required)]...)
	}

	return liveSegments
}

// hlsHeader generates the header string for an HLS playlist with the specified target duration.
func hlsHeader(dur float64) string {
	return "#EXTM3U\n" +
		"#EXT-X-VERSION:3\n" +
		"#EXT-X-TARGETDURATION:" + strconv.FormatFloat(dur, 'f', -1, 64) + "\n" +
		"#EXT-X-MEDIA-SEQUENCE:0\n"
}

// hlsSegment generates an HLS segment entry with the specified duration and path.
func hlsSegment(dur float64, path string) string {
	duration := strconv.FormatFloat(dur, 'f', -1, 64)
	return "#EXTINF:" + duration + ",\n" + path + "\n"
}
