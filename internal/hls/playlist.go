// Package hls provides functionality for handling HTTP Live Streaming (HLS) playlists and segments.
package hls

import (
	"math"
	"strconv"
	"time"
)

// Playlist represents an HLS playlist structure.
type Playlist struct {
	LiveSegmentsAmount int // The number of live segments in the playlist.
	MaxSegmentDuration int // The maximum duration (in seconds) of a segment in the playlist.

	mediaSequence        int64
	disconSequence       int64
	lastDisconUpdate     time.Time
	currentTrackSegments []*Segment
	nextTrackSegments    []*Segment
	currentSegmentPath   string
}

// NewPlaylist creates and returns a new Playlist instance with the provided current and next track segments.
// It initializes the playlist with default values for live segments amount, max segment duration, media sequence,
// discontinuity sequence, and last discontinuity update time.
//
// Parameters:
//   - cur: The list of segments for the current track.
//   - next: The list of segments for the next track.
//
// Returns:
//   - A pointer to the newly created Playlist instance.
func NewPlaylist(cur, next []*Segment) *Playlist {
	return &Playlist{
		LiveSegmentsAmount: DefaultLiveSegmentsAmount,
		MaxSegmentDuration: DefaultMaxSegmentDuration,

		mediaSequence:    0,
		disconSequence:   0,
		lastDisconUpdate: time.Now(),

		currentTrackSegments: cur,
		nextTrackSegments:    next,

		currentSegmentPath: "",
	}
}

// Generate constructs and returns the HLS playlist as a string based on the elapsed time.
// It updates the discontinuity sequence, calculates the starting segment index, collects live segments,
// and formats them into the HLS playlist format.
//
// Parameters:
//   - elapsedTime: The elapsed time in seconds used to determine the current segment index.
//
// Returns:
//   - A string representing the generated HLS playlist.
func (p *Playlist) Generate(elapsedTime float64) string {
	offset := math.Mod(elapsedTime, float64(p.MaxSegmentDuration))
	liveSegments := p.currentSegments(elapsedTime)
	prevSegmentPath := p.currentSegmentPath

	if len(liveSegments) > 0 {
		p.currentSegmentPath = liveSegments[0].Path
	}

	p.UpdateDisconSequence(elapsedTime)
	if prevSegmentPath != p.currentSegmentPath {
		p.mediaSequence++
	}

	playlist := hlsHeader(p.MaxSegmentDuration, p.mediaSequence, p.disconSequence, offset)
	for _, seg := range liveSegments {
		playlist += hlsSegment(seg.Duration, seg.Path, seg.IsFirst)
	}

	return playlist
}

// Next updates the playlist by moving the next track segments to the current track segments
// and assigning the provided segments as the new next track segments.
//
// Parameters:
//   - next: The new list of segments to be set as the next track segments.
func (p *Playlist) Next(next []*Segment) {
	p.currentTrackSegments = p.nextTrackSegments
	p.nextTrackSegments = next
}

// ChangeNext replays segments for the next track.
//
// Parameters:
//   - next: The new list of segments to be set as the next track segments.
func (p *Playlist) ChangeNext(next []*Segment) {
	p.nextTrackSegments = next
}

// AddSegments appends the provided segments to the next track segments list.
//
// Parameters:
//   - segments: The list of segments to append to the next track segments.
func (p *Playlist) AddSegments(segments []*Segment) {
	p.nextTrackSegments = append(p.nextTrackSegments, segments...)
}

// SetMediaSequence set a new sequence number for mediaSequence.
func (p *Playlist) SetMediaSequence(sequence int64) {
	p.mediaSequence = sequence
}

// UpdateDisconSequence updates the discontinuity sequence if a discontinuity is detected.
//
// Parameters:
//   - elapsedTime: The elapsed time in seconds used to calculate the current segment index.
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

func (p *Playlist) FirstNextTrackSegment() *Segment {
	if len(p.nextTrackSegments) > 0 {
		return p.nextTrackSegments[0]
	}

	return nil
}

// currentSegments gathers enough segments from current and next tracks to meet liveSegmentsAmount
func (p *Playlist) currentSegments(elapsedTime float64) []*Segment {
	startIndex := p.calcCurrentSegmentIndex(elapsedTime)
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

func (p *Playlist) calcCurrentSegmentIndex(elapsedTime float64) int {
	return int(math.Floor(elapsedTime / float64(p.MaxSegmentDuration)))
}

// hlsHeader generates the header string for an HLS playlist with the specified target duration.
func hlsHeader(dur int, mediaSeq, disconSeq int64, offset float64) string {
	currentTime := time.Now().UTC().Round(time.Millisecond).Format(timeFormat)
	return "#EXTM3U\n" +
		"#EXT-X-VERSION:6\n" +
		"#EXT-X-PROGRAM-DATE-TIME:" + currentTime + "\n" +
		"#EXT-X-TARGETDURATION:" + strconv.Itoa(dur) + "\n" +
		"#EXT-X-MEDIA-SEQUENCE:" + strconv.FormatInt(mediaSeq, 10) + "\n" +
		"#EXT-X-DISCONTINUITY-SEQUENCE:" + strconv.FormatInt(disconSeq, 10) + "\n" +
		"#EXT-X-START:TIME-OFFSET=" + strconv.FormatFloat(offset, 'f', 2, 64) + "\n"
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
