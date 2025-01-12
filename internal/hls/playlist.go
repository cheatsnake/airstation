package hls

import (
	"math"
	"strconv"
	"time"
)

type Playlist struct {
	LiveSegmentsAmount int
	MaxSegmentDuration int

	mediaSequence    int64
	disconSequence   int64
	lastDisconUpdate time.Time

	currentTrackSegments []*Segment
	nextTrackSegments    []*Segment
}

func NewPlaylist(cur, next []*Segment, maxDuration int, liveAmount int) *Playlist {
	return &Playlist{
		LiveSegmentsAmount: liveAmount,
		MaxSegmentDuration: maxDuration,

		mediaSequence:    0,
		disconSequence:   0,
		lastDisconUpdate: time.Now(),

		currentTrackSegments: cur,
		nextTrackSegments:    next,
	}
}

func (p *Playlist) Generate(elapsedTime float64) string {
	p.UpdateDisconSequence(elapsedTime)

	playlist := hlsHeader(p.MaxSegmentDuration, p.mediaSequence, p.disconSequence)
	firstSegmentIndex := p.calcCurrentSegmentIndex(elapsedTime)
	liveSegments := p.collectLiveSegments(firstSegmentIndex)

	for _, seg := range liveSegments {
		playlist += hlsSegment(seg.Duration, seg.Path, seg.IsFirst)
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

func (p *Playlist) UpdateMediaSequence() {
	p.mediaSequence++
}

func (p *Playlist) UpdateDisconSequence(elapsedTime float64) {
	elapsedFromLastUpdate := time.Until(p.lastDisconUpdate).Seconds()
	if math.Abs(elapsedFromLastUpdate) < float64(p.MaxSegmentDuration) {
		return
	}

	index := p.calcCurrentSegmentIndex(elapsedTime)

	// if the current track segment is the second and it is not the very first track,
	// there was a discontinuty, so we increment the discontinuty counter
	if index == 1 && p.mediaSequence > 1 {
		p.disconSequence++
		p.lastDisconUpdate = time.Now()
	}
}

func (p *Playlist) calcCurrentSegmentIndex(elapsedTime float64) int {
	return int(math.Floor(elapsedTime / float64(p.MaxSegmentDuration)))
}

// collectLiveSegments gathers enough segments from current and next tracks to meet liveSegmentsAmount
func (p *Playlist) collectLiveSegments(startIndex int) []*Segment {
	liveSegments := make([]*Segment, 0, p.LiveSegmentsAmount)

	if startIndex < len(p.currentTrackSegments) {
		endIndex := startIndex + p.LiveSegmentsAmount
		if endIndex >= len(p.currentTrackSegments) {
			endIndex = len(p.currentTrackSegments)
		}

		liveSegments = append(liveSegments, p.currentTrackSegments[startIndex:endIndex]...)
	}

	if len(liveSegments) < p.LiveSegmentsAmount {
		required := p.LiveSegmentsAmount - len(liveSegments)
		liveSegments = append(liveSegments, p.nextTrackSegments[:min(len(p.nextTrackSegments), required)]...)
	}

	return liveSegments
}

// hlsHeader generates the header string for an HLS playlist with the specified target duration.
func hlsHeader(dur int, mediaSeq, disconSeq int64) string {
	return "#EXTM3U\n" +
		"#EXT-X-VERSION:3\n" +
		"#EXT-X-TARGETDURATION:" + strconv.Itoa(dur) + "\n" +
		"#EXT-X-MEDIA-SEQUENCE:" + strconv.FormatInt(mediaSeq, 10) + "\n" +
		"#EXT-X-DISCONTINUITY-SEQUENCE:" + strconv.FormatInt(disconSeq, 10) + "\n"
}

// hlsSegment generates an HLS segment entry with the specified duration and path.
func hlsSegment(dur float64, path string, isDiscon bool) string {
	disconTag := ""

	if isDiscon {
		disconTag = "#EXT-X-DISCONTINUITY\n"
	}

	duration := strconv.FormatFloat(dur, 'f', 2, 64)
	return disconTag +
		"#EXTINF:" + duration + ",\n" +
		path + "\n"
}
